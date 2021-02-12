use std::{any::Any, collections::{HashMap, HashSet}, fs::File, ops::Deref, path::{Path, PathBuf}, rc::Rc};
use std::hash::Hash;

use dashmap::DashMap;
use futures::{executor::block_on, lock::Mutex};
use log::{debug, trace};
use tonic::{IntoRequest, transport::Channel};
use anyhow::{anyhow, Context, Result};

use crate::{core_api_bindings::api_container_api::{RegisterServiceArgs, RemoveServiceArgs, StartServiceArgs, test_execution_service_client::TestExecutionServiceClient, test_execution_service_server::TestExecutionService}, services::{availability_checker::AvailabilityChecker, docker_container_initializer::DockerContainerInitializer, service::Service, service_wrapper::ServiceInterfaceWrapper}};

use super::network::Network;

// TODO Make a type
const DEFAULT_PARTITION_ID: &str = "";

// This value - where the suite execution volume will be mounted on the testsuite container - is
//  hardcoded inside Kurtosis Core
const SUITE_EX_VOL_MOUNTPOINT: &str = "/suite-execution";

struct ServiceInfo {
	ip_addr: String,
	service_interface_wrapper: Box<dyn Fn(&str, &str) -> Box<dyn Service>>,
}

pub struct NetworkContext {
    client: TestExecutionServiceClient<Channel>,
	// TODO Make key a separate FilesArtifactID type
	files_artifact_urls: HashMap<String, String>,
    all_service_info: DashMap<String, ServiceInfo>,
}

impl NetworkContext {
    pub fn new(client: TestExecutionServiceClient<Channel>, files_artifact_urls: HashMap<String, String>) -> NetworkContext {
        return NetworkContext {
            client,
			files_artifact_urls,
            all_service_info: DashMap::new(),
        };
    }

    pub fn add_service<S: Service>(&mut self, service_id: &str, initializer: &dyn DockerContainerInitializer<S>) -> Result<(Box<S>, AvailabilityChecker)> {
        let (service_ptr, availability_checker) = self.add_service_to_partition(service_id, DEFAULT_PARTITION_ID, initializer)
			.context(format!("An error occurred adding service '{}' to the network in the default partition", service_id))?;
		return Ok((service_ptr, availability_checker));
	}

    pub fn add_service_to_partition<'a, S: Service>(&mut self, service_id: &str, partition_id: &str, initializer: &dyn DockerContainerInitializer<S>) -> Result<(Box<S>, AvailabilityChecker)> {
		trace!("Registering new service ID with Kurtosis API...");
		let files_to_generate = NetworkContext::convert_hashset_to_hashmap(initializer.get_files_to_mount());
		let args = RegisterServiceArgs{
		    service_id: service_id.to_owned(),
		    partition_id: partition_id.to_owned(),
		    files_to_generate,
		};
		let register_service_args = tonic::Request::new(args);
		let register_service_resp = block_on(self.client.register_service(register_service_args))
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
		initializer.initialize_mounted_files(generated_files_fps)
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
		let start_cmd_args_opt = initializer.get_start_command(generated_files_abs_filepaths_on_service, &service_ip_addr)
			.context("Failed to create start command")?;
		trace!("Successfully created start command for service");

		trace!("Starting new service with Kurtosis API...");
		let start_service_args = StartServiceArgs{
		    service_id: service_id.to_owned(),
		    docker_image: initializer.get_docker_image().to_owned(),
		    used_ports: NetworkContext::convert_hashset_to_hashmap(initializer.get_used_ports()),
			// NOTE: If empty vector isn't the "use default Docker CMD" then we need something else
		    start_cmd_args: start_cmd_args_opt.unwrap_or(Vec::new()),
		    docker_env_vars: HashMap::new(),  // TODO actually support Docker env vars!
		    suite_execution_vol_mnt_dirpath: initializer.get_test_volume_mountpoint().to_owned(),
		    files_artifact_mount_dirpaths: artifact_url_to_mount_dirpath,
		};
		let start_service_req = tonic::Request::new(start_service_args);
		let start_service_resp = block_on(self.client.start_service(start_service_req))
			.context("An error occurred starting the service with the Kurtosis API")?
			.into_inner();
		trace!("Successfully started service with Kurtosis API");

		trace!("Creating service interface...");
		let service_interface_wrapper = initializer.get_service_wrapping_func();
		let result_service_ptr = NetworkContext::call_service_interface_wrapping_func(
			initializer.get_service_wrapping_func(), 
			service_id, 
			&service_ip_addr
		);
		let casted_result_service_ptr_or_err = result_service_ptr.downcast::<S>();
		let casted_result_service_ptr: Box<S>;
		match casted_result_service_ptr_or_err {
			Ok(ptr) => casted_result_service_ptr = ptr,
			Err(_) => return Err(anyhow!(
				"An error occurred casting service to appropriate type"
			)),
		}

		// Yes, this creates a second Service interfaces; yes, this is inefficient; yes, there's probably a better way to do this
		// At this point though I'm *so* sick of fighting Rust's compiler and its un-Google-able, seemingly-draconian errors so
		// I'm doing it this way (e.g. - how can I both store a value in a hashmap, use it in the AvailabilityChecker, and return
		// it to the user??) ~ ktoday, 2021-02-12
		let availability_checker_service_ptr = NetworkContext::call_service_interface_wrapping_func(
			initializer.get_service_wrapping_func(), 
			service_id, 
			&service_ip_addr
		);
		let availability_checker = AvailabilityChecker::new(service_id, availability_checker_service_ptr);
		trace!("Successfully created service interface");

		let new_service_info = ServiceInfo{
		    ip_addr: service_ip_addr,
		    service_interface_wrapper: service_interface_wrapper,
		};
		self.all_service_info.insert(service_id.to_owned(), new_service_info);

		return Ok((casted_result_service_ptr, availability_checker));
    }

    pub fn get_service<S: Service>(&self, service_id: &str) -> Result<Box<S>> {
		let desired_service_info = self.all_service_info.get(service_id)
			.context(format!("No service found with ID '{}'", service_id))?;
		let service_interface_wrapper_func = &desired_service_info.service_interface_wrapper;
		let service_ptr = NetworkContext::call_service_interface_wrapping_func(
			service_interface_wrapper_func, 
			service_id, 
			&desired_service_info.ip_addr,
		);
		let casted_service_ptr_or_err = service_ptr.downcast::<S>();
		let result: Box<S>;
		match casted_service_ptr_or_err {
			Ok(casted_ptr) => result = casted_ptr,
			Err(_) => return Err(anyhow!(
				"An error occurred casting service with ID '{}' to appropriate type",
				service_id,
			)),
		}
		return Ok(result);
    }

    pub fn remove_service(&mut self, service_id: &str, container_stop_timeout_seconds: u64) -> Result<()> {
		debug!("Removing service '{}'...", service_id);
		let args = RemoveServiceArgs{
		    service_id: service_id.to_owned(),
			// NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
			//  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
			// Independent of adding/removing them from the network
		    container_stop_timeout_seconds,
		};
		let req = tonic::Request::new(args);
		block_on(self.client.remove_service(req))
			.context(format!("An error occurred removing service '{}' from the network", service_id))?;
		self.all_service_info.remove(service_id);
		debug!("Successfully removed service ID '{}'", service_id);
		return Ok(());
	}

    fn get_repartition_builder() {
        /*
        func (networkCtx NetworkContext) GetRepartitionerBuilder(isDefaultPartitionConnectionBlocked bool) *RepartitionerBuilder {
	// This function doesn't need a mutex lock because (as of 2020-12-28) it doesn't touch internal state whatsoever
	return newRepartitionerBuilder(isDefaultPartitionConnectionBlocked)
}
 */
    }

    fn repartition_network() {
        /* 	networkCtx.mutex.Lock()
	defer networkCtx.mutex.Unlock()

	partitionServices := map[string]*core_api_bindings.PartitionServices{}
	for partitionId, serviceIdSet := range repartitioner.partitionServices {
		serviceIdStrPseudoSet := map[string]bool{}
		for _, serviceId := range serviceIdSet.getElems() {
			serviceIdStr := string(serviceId)
			serviceIdStrPseudoSet[serviceIdStr] = true
		}
		partitionIdStr := string(partitionId)
		partitionServices[partitionIdStr] = &core_api_bindings.PartitionServices{
			ServiceIdSet: serviceIdStrPseudoSet,
		}
	}

	partitionConns := map[string]*core_api_bindings.PartitionConnections{}
	for partitionAId, partitionAConnsMap := range repartitioner.partitionConnections {
		partitionAConnsStrMap := map[string]*core_api_bindings.PartitionConnectionInfo{}
		for partitionBId, connInfo := range partitionAConnsMap {
			partitionBIdStr := string(partitionBId)
			partitionAConnsStrMap[partitionBIdStr] = connInfo
		}
		partitionAConns := &core_api_bindings.PartitionConnections{
			ConnectionInfo: partitionAConnsStrMap,
		}
		partitionAIdStr := string(partitionAId)
		partitionConns[partitionAIdStr] = partitionAConns
	}

	repartitionArgs := &core_api_bindings.RepartitionArgs{
		PartitionServices:    partitionServices,
		PartitionConnections: partitionConns,
		DefaultConnection:    repartitioner.defaultConnection,
	}
	if _, err := networkCtx.client.Repartition(context.Background(), repartitionArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return nil
}
 */
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

	fn call_service_interface_wrapping_func<F>(
		func: F, 
		service_id: &str, 
		ip_addr: &str
	) -> Box<dyn Service> where F: Fn(&str, &str) -> Box<dyn Service> {
		return func(service_id, ip_addr);
	}
}
