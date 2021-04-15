use std::{collections::{HashMap, HashSet}, fs::File, ops::{Deref, DerefMut}, path::{PathBuf}, rc::Rc};
use std::hash::Hash;

use dashmap::DashMap;
use log::{debug, trace};
use tokio::runtime::Runtime;
use tonic::{transport::Channel};
use anyhow::{anyhow, Context, Result};

use crate::{core_api_bindings::api_container_api::{PartitionConnectionInfo, PartitionConnections, PartitionServices, RegisterServiceArgs, RemoveServiceArgs, RepartitionArgs, StartServiceArgs, test_execution_service_client::TestExecutionServiceClient}, services::{availability_checker::AvailabilityChecker, container_config_factory::ContainerConfigFactory, service::{Service, ServiceId}, service_context::ServiceContext}};

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
    pub fn add_service<S: Service>(&mut self, service_id: &ServiceId, config_factory: &dyn ContainerConfigFactory<S>) -> Result<(Rc<S>, AvailabilityChecker)> {
		let (service_ptr, availability_checker) = self.add_service_to_partition(service_id, &DEFAULT_PARTITION_ID_STR.to_owned(), config_factory)
			.context(format!("An error occurred adding service '{}' to the network in the default partition", service_id))?;
		return Ok((service_ptr, availability_checker));
	}

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn add_service_to_partition<S: Service>(&mut self, service_id: &ServiceId, partition_id: &PartitionId, config_factory: &dyn ContainerConfigFactory<S>) -> Result<(Rc<S>, AvailabilityChecker)> {
		trace!("Registering new service ID with Kurtosis API...");
		let args = RegisterServiceArgs{
		    service_id: service_id.to_owned(),
		    partition_id: partition_id.to_owned(),
		};
		let register_service_args = tonic::Request::new(args);
		let register_service_resp = self.async_runtime.block_on(self.client.register_service(register_service_args))
			.context(format!("An error occurred registering service with ID '{}' with the Kurtosis API", service_id))?
			.into_inner();
		let service_ip_addr = register_service_resp.ip_addr;
		let container_creation_config = config_factory.get_creation_config(&service_ip_addr)
			.context("An error occurred getting the container creation config")?;
		let service_context_client = self.client.clone();
		let service_context = ServiceContext::new(
			self.async_runtime.clone(), 
			service_context_client, 
			service_id.to_owned(), 
			service_ip_addr.to_owned(),
			SUITE_EX_VOL_MOUNTPOINT.to_owned(),
			container_creation_config.get_test_volume_mountpoint().to_owned(),
		);
		trace!("New service successfully registered with Kurtosis API");

		trace!("Initializing generated files in suite execution volume...");
		let mut files_to_generate: HashSet<String> = HashSet::new();
		for (file_id, _) in container_creation_config.get_file_generating_funcs() {
			files_to_generate.insert(file_id.to_owned());
		}
		let generated_file_filepaths = service_context.generate_files(files_to_generate)
			.context(format!("An error occurred generating the files needed for service startup"))?;
		let mut generated_files_abs_filepaths_on_service: HashMap<String, PathBuf> = HashMap::new();
		for (file_id, initializing_func_arc_mutex) in container_creation_config.get_file_generating_funcs() {
			let filepaths = generated_file_filepaths.get(file_id)
				.context(format!(
					"Needed to initialize file for file ID '{}', but no generated file filepaths were found for that file ID; this is a Kurtosis bug",
					file_id,
				))?;
			let fp = File::create(&filepaths.absolute_filepath_on_testsuite_container)
				.context(format!("An error occurred opening file pointer for file '{}'", file_id))?;
			let mut initializing_func = initializing_func_arc_mutex.lock().unwrap();
			initializing_func.deref_mut()(fp)
				.context(format!("The function to initialize file with ID '{}' returned an error", file_id))?;
			generated_files_abs_filepaths_on_service.insert(file_id.to_owned(), filepaths.absolute_filepath_on_service_container.clone());
		}
		trace!("Successfully initialized generated files in suite execution volume");

		let container_run_config = config_factory.get_run_config(&service_ip_addr, generated_files_abs_filepaths_on_service)
			.context("An error occurred getting the container run config")?;

		trace!("Creating files artifact URL -> mount dirpaths map...");
		let mut artifact_url_to_mount_dirpath: HashMap<String, String> = HashMap::new();
		for (files_artifact_id, mount_dirpath) in container_creation_config.get_files_artifact_mountpoints() {
			let artifact_url = self.files_artifact_urls.get(files_artifact_id)
				.context(format!(
					"Service requested file artifact '{}', but the network context doesn't have a URL for that file artifact; this is a bug with Kurtosis itself",
					files_artifact_id
				))?;
			artifact_url_to_mount_dirpath.insert(artifact_url.to_owned(), mount_dirpath.to_owned());
		}

		trace!("Successfully created files artifact URL -> mount dirpaths map");

		trace!("Starting new service with Kurtosis API...");
		let start_service_args = StartServiceArgs{
		    service_id: service_id.to_owned(),
		    docker_image: container_creation_config.get_image().to_owned(),
		    used_ports: NetworkContext::convert_hashset_to_hashmap(container_creation_config.get_used_ports().to_owned()),
			entrypoint_args: container_run_config.get_entrypoint_override_args().to_owned(),
			cmd_args: container_run_config.get_cmd_override_args().to_owned(),
		    docker_env_vars: container_run_config.get_environment_variable_overrides().to_owned(),
		    suite_execution_vol_mnt_dirpath: container_creation_config.get_test_volume_mountpoint().to_owned(),
		    files_artifact_mount_dirpaths: artifact_url_to_mount_dirpath,
		};
		let start_service_req = tonic::Request::new(start_service_args);
		self.async_runtime.block_on(self.client.start_service(start_service_req))
			.context("An error occurred starting the service with the Kurtosis API")?
			.into_inner();
		trace!("Successfully started service with Kurtosis API");


		trace!("Creating service interface...");
		let result_service_ptr = container_creation_config.get_service_creating_func()(service_context);
		let casted_result_service_rc = Rc::new(result_service_ptr);
		trace!("Successfully created service interface");

		self.all_services.insert(service_id.to_owned(), casted_result_service_rc.clone());

		let availability_checker = AvailabilityChecker::new(service_id, casted_result_service_rc.clone());

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