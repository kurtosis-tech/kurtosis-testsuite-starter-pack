use std::{collections::{HashMap, HashSet}, fs::File, ops::Deref, path::{PathBuf}, rc::Rc};
use std::hash::Hash;

use dashmap::DashMap;
use log::{debug, trace};
use tokio::runtime::Runtime;
use tonic::{transport::Channel};
use anyhow::{anyhow, Context, Result};

use crate::{core_api_bindings::api_container_api::{PartitionConnectionInfo, PartitionConnections, PartitionServices, RegisterServiceArgs, RemoveServiceArgs, RepartitionArgs, StartServiceArgs, test_execution_service_client::TestExecutionServiceClient}, services::{availability_checker::AvailabilityChecker, docker_container_initializer::DockerContainerInitializer, service::{Service, ServiceId}, service_context::ServiceContext}};

use super::network::Network;

// TODO Make a type
const DEFAULT_PARTITION_ID_STR: &str = "";

// This value - where the suite execution volume will be mounted on the testsuite container - is
//  hardcoded inside Kurtosis Core
const SUITE_EX_VOL_MOUNTPOINT: &str = "/suite-execution";

pub type PartitionId = String;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct NetworkContext {
	async_runtime: Rc<Runtime>,
    client: TestExecutionServiceClient<Channel>,
	// TODO Make key a separate FilesArtifactID type
	files_artifact_urls: HashMap<String, String>,
    all_services: DashMap<ServiceId, Rc<dyn Service>>,
}

impl NetworkContext {
    pub fn new(async_runtime: Rc<Runtime>, client: TestExecutionServiceClient<Channel>, files_artifact_urls: HashMap<String, String>) -> NetworkContext {
        return NetworkContext {
			async_runtime,
            client,
			files_artifact_urls,
            all_services: DashMap::new(),
        };
    }

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn add_service<S: Service>(&mut self, service_id: &ServiceId, initializer: &dyn DockerContainerInitializer<S>) -> Result<(Rc<S>, AvailabilityChecker)> {
		let (service_ptr, availability_checker) = self.add_service_to_partition(service_id, &DEFAULT_PARTITION_ID_STR.to_owned(), initializer)
			.context(format!("An error occurred adding service '{}' to the network in the default partition", service_id))?;
		return Ok((service_ptr, availability_checker));
	}

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn add_service_to_partition<S: Service>(&mut self, service_id: &ServiceId, partition_id: &PartitionId, initializer: &dyn DockerContainerInitializer<S>) -> Result<(Rc<S>, AvailabilityChecker)> {
		trace!("Registering new service ID with Kurtosis API...");
		let files_to_generate = NetworkContext::convert_hashset_to_hashmap(initializer.get_files_to_generate());
		let args = RegisterServiceArgs{
		    service_id: service_id.to_owned(),
		    partition_id: partition_id.to_owned(),
		    files_to_generate,
		};
		let register_service_args = tonic::Request::new(args);
		let register_service_resp = self.async_runtime.block_on(self.client.register_service(register_service_args))
			.context(format!("An error occurred registering service with ID '{}' with the Kurtosis API", service_id))?
			.into_inner();
		
		let suite_ex_vol_mountpoint_on_service = initializer.get_test_volume_mountpoint();
		let generated_files_relative_filepaths = register_service_resp.generated_files_relative_filepaths;
		let mut generated_files_fps: HashMap<String, File> = HashMap::new();
		let mut generated_files_abs_filepaths_on_service: HashMap<String, PathBuf> = HashMap::new();
		for (file_id, relative_filepath) in generated_files_relative_filepaths {
			// Per https://users.rust-lang.org/t/what-is-the-idiomatic-way-to-create-a-path-from-str-fragments/42882/2 , 
			// this is the best way to join multiple fragments into a single path
			let abs_filepath_on_testsuite: PathBuf = [SUITE_EX_VOL_MOUNTPOINT, &relative_filepath].iter().collect();
			debug!("Opening generated file at '{}' for writing...", abs_filepath_on_testsuite.display());
			let fp = File::create(abs_filepath_on_testsuite)
				.context(format!("Could not open generated file '{}' for writing", file_id))?;
			generated_files_fps.insert(file_id.clone(), fp);

			let abs_filepath_on_service: PathBuf = [suite_ex_vol_mountpoint_on_service, &relative_filepath].iter().collect();
			generated_files_abs_filepaths_on_service.insert(file_id.clone(), abs_filepath_on_service);
		}

		trace!("Initializing generated files...");
		initializer.initialize_generated_files(generated_files_fps)
			.context("An error occurred initializing the generated files")?;
		trace!("Successfully initialized generated files");

		trace!("Creating files artifact URL -> mount dirpaths map...");
		let mut artifact_url_to_mount_dirpath: HashMap<String, String> = HashMap::new();
		for (files_artifact_id, mount_dirpath) in initializer.get_files_artifact_mountpoints() {
			let artifact_url = self.files_artifact_urls.get(&files_artifact_id)
				.context(format!(
					"Service requested file artifact '{}', but the network context doesn't have a URL for that file artifact; this is a bug with Kurtosis itself",
					files_artifact_id
				))?;
			artifact_url_to_mount_dirpath.insert(artifact_url.to_owned(), mount_dirpath);
		}

		trace!("Successfully created files artifact URL -> mount dirpaths map");

		trace!("Creating start command for service...");
		let service_ip_addr = register_service_resp.ip_addr;
		let (entrypoint_args_opt, cmd_args_opt) = initializer.get_start_command_overrides(generated_files_abs_filepaths_on_service, &service_ip_addr)
			.context("Failed to get start command overrides")?;
		trace!("Successfully created start command for service");

		trace!("Starting new service with Kurtosis API...");
		let start_service_args = StartServiceArgs{
		    service_id: service_id.to_owned(),
		    docker_image: initializer.get_docker_image().to_owned(),
		    used_ports: NetworkContext::convert_hashset_to_hashmap(initializer.get_used_ports()),
			entrypoint_args: entrypoint_args_opt.unwrap_or(Vec::new()), // Empty vector says "don't override anything"
			cmd_args: cmd_args_opt.unwrap_or(Vec::new()), // Empty vector says "don't override anything"
		    docker_env_vars: HashMap::new(),  // TODO actually support Docker env vars!
		    suite_execution_vol_mnt_dirpath: initializer.get_test_volume_mountpoint().to_owned(),
		    files_artifact_mount_dirpaths: artifact_url_to_mount_dirpath,
		};
		let start_service_req = tonic::Request::new(start_service_args);
		self.async_runtime.block_on(self.client.start_service(start_service_req))
			.context("An error occurred starting the service with the Kurtosis API")?
			.into_inner();
		trace!("Successfully started service with Kurtosis API");

		let service_context_client = self.client.clone();
		let service_context = ServiceContext::new(self.async_runtime.clone(), service_context_client, service_id.to_owned(), service_ip_addr);

		trace!("Creating service interface...");
		let result_service_ptr = initializer.get_service(service_context);
		let casted_result_service_rc = Rc::new(*result_service_ptr);
		let availability_checker = AvailabilityChecker::new(service_id, casted_result_service_rc.clone());
		trace!("Successfully created service interface");

		self.all_services.insert(service_id.to_owned(), casted_result_service_rc.clone());

		return Ok((casted_result_service_rc, availability_checker));
    }

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn get_service<S: Service>(&self, service_id: &ServiceId) -> Result<Rc<S>> {
		let service_ptr_ptr = self.all_services.get(service_id)
			.context(format!("No service found with ID '{}'", service_id))?;
		let service_ptr = service_ptr_ptr.deref().clone();
		let casted_service_ptr_or_err = service_ptr.downcast_rc::<S>();
		let result: Rc<S>;
		match casted_service_ptr_or_err {
			Ok(casted_ptr) => result = casted_ptr,
			Err(_) => return Err(anyhow!(
				"An error occurred casting service with ID '{}' to appropriate type",
				service_id,
			)),
		}
		return Ok(result);
    }

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn remove_service(&mut self, service_id: &ServiceId, container_stop_timeout_seconds: u64) -> Result<()> {
		debug!("Removing service '{}'...", service_id);
		let args = RemoveServiceArgs{
		    service_id: service_id.to_owned(),
			// NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
			//  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
			// Independent of adding/removing them from the network
		    container_stop_timeout_seconds,
		};
		let req = tonic::Request::new(args);
		self.async_runtime.block_on(self.client.remove_service(req))
			.context(format!("An error occurred removing service '{}' from the network", service_id))?;
		self.all_services.remove(service_id);
		debug!("Successfully removed service ID '{}'", service_id);
		return Ok(());
	}

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn repartition_network(
		&mut self, 
		partition_services: HashMap<PartitionId, HashSet<ServiceId>>,
		partition_connections: HashMap<PartitionId, HashMap<PartitionId, PartitionConnectionInfo>>,
		default_connection_info: PartitionConnectionInfo,
	) -> Result<()> {
		let mut req_partition_services: HashMap<String, PartitionServices> = HashMap::new();
		for (partition_id, service_id_set) in partition_services {
			let mut service_id_str_pseudo_set: HashMap<String, bool> = HashMap::new();
			for service_id in service_id_set {
				service_id_str_pseudo_set.insert(service_id, true);
			}
			req_partition_services.insert(partition_id, PartitionServices{
			    service_id_set: service_id_str_pseudo_set,
			});
		}

		let mut req_partition_connections: HashMap<String, PartitionConnections> = HashMap::new();
		for (partition_a_id, partition_a_conns_map) in partition_connections {
			let mut partition_a_conns_str_map: HashMap<String, PartitionConnectionInfo> = HashMap::new();
			for (partition_b_id, conn_info) in partition_a_conns_map {
				partition_a_conns_str_map.insert(partition_b_id, conn_info);
			}
			let partition_a_conns = PartitionConnections{
			    connection_info: partition_a_conns_str_map,
			};
			req_partition_connections.insert(partition_a_id, partition_a_conns);
		}

		let args = RepartitionArgs{
		    partition_services: req_partition_services,
		    partition_connections: req_partition_connections,
			// NOTE: It's unclear why tonic generates this as "optional"; it's not optional in the .proto file
		    default_connection: Some(default_connection_info),
		};

		let req = tonic::Request::new(args);
		self.async_runtime.block_on(self.client.repartition(req))
			.context("An error occurred repartitioning the test network")?;
		return Ok(());
    }

	fn convert_hashset_to_hashmap<T>(
		set: HashSet<T>
	) -> HashMap<T, bool> where T: Eq + Hash {
		let mut result: HashMap<T, bool> = HashMap::new();
		for elem in set {
			result.insert(elem, true);
		}
		return result;
	}
}

impl Network for NetworkContext { }