use std::collections::HashMap;

use futures::lock::Mutex;
use tonic::transport::Channel;

use crate::{core_api_bindings::api_container_api::{test_execution_service_client::TestExecutionServiceClient, test_execution_service_server::TestExecutionService}, services::{docker_container_initializer::DockerContainerInitializer, service::Service}};

// TODO Make a type
const DEFAULT_PARTITION_ID: &str = "";

pub struct NetworkContext {
    client: TestExecutionServiceClient<Channel>,
    mutex: Mutex<HashMap<String, Box<dyn Service>>>,
}

impl NetworkContext {
    pub fn new(client: TestExecutionServiceClient<Channel>) -> NetworkContext {
        return NetworkContext {
            client,
            mutex: Mutex::new(HashMap::new()),
        };
    }

    pub fn add_service(&self, service_id: &str, initializer: &dyn DockerContainerInitializer) -> Rust<(&dyn Service, &dyn AvailabilityCh{
        self.add_service_to_partition(service_id, DEFAULT_PARTITION_ID)
        /* 
        func (networkCtx *NetworkContext) AddService(
		serviceId services.ServiceID,
		initializer services.DockerContainerInitializer) (services.Service, services.AvailabilityChecker, error) {
	// Go mutexes aren't re-entrant, so we lock the mutex inside this call
	service, availabilityChecker, err := networkCtx.AddServiceToPartition(
		serviceId,
		defaultPartitionId,
		initializer)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to the network in the default partition", serviceId)
	}

	return service, availabilityChecker, nil
}
*/

    }

    pub fn add_service_to_partition(&self, service_id: &str, partition_id: &str, initializer: &dyn DockerContainerInitializer) {
/*
func (networkCtx *NetworkContext) AddServiceToPartition(
		serviceId services.ServiceID,
		partitionId PartitionID,
		initializer services.DockerContainerInitializer) (services.Service, services.AvailabilityChecker, error) {
	networkCtx.mutex.Lock()
	defer networkCtx.mutex.Unlock()

	ctx := context.Background()

	logrus.Tracef("Registering new service ID with Kurtosis API...")
	registerServiceArgs := &core_api_bindings.RegisterServiceArgs{
		ServiceId:       string(serviceId),
		PartitionId:     string(partitionId),
		FilesToGenerate: initializer.GetFilesToMount(),
	}
	registerServiceResp, err := networkCtx.client.RegisterService(ctx, registerServiceArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred registering service with ID '%v' with the Kurtosis API",
			serviceId)
	}
	logrus.Tracef("New service successfully registered with Kurtosis API")

	suiteExVolMountpointOnService := initializer.GetTestVolumeMountpoint()
	generatedFilesRelativeFilepaths := registerServiceResp.GeneratedFilesRelativeFilepaths
	generatedFilesFps := map[string]*os.File{}
	generatedFilesAbsoluteFilepathsOnService := map[string]string{}
	for fileId, relativeFilepath := range generatedFilesRelativeFilepaths {
		absoluteFilepathOnTestsuite := path.Join(test_suite_container_mountpoints.SuiteExVolMountpoint, relativeFilepath)
		logrus.Debugf("Opening generated file at '%v' for writing...", absoluteFilepathOnTestsuite)
		fp, err := os.Create(absoluteFilepathOnTestsuite)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"Could not open generated file '%v' for writing",
				fileId)
		}
		defer fp.Close()
		generatedFilesFps[fileId] = fp

		absoluteFilepathOnService := path.Join(suiteExVolMountpointOnService, relativeFilepath)
		generatedFilesAbsoluteFilepathsOnService[fileId] = absoluteFilepathOnService
	}

	logrus.Trace("Initializing generated files...")
	if err := initializer.InitializeMountedFiles(generatedFilesFps); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred initializing the generated files")
	}
	logrus.Trace("Successfully initialized generated files")


	logrus.Tracef("Creating files artifact URL -> mount dirpaths map...")
	artifactUrlToMountDirpath := map[string]string{}
	for filesArtifactId, mountDirpath := range initializer.GetFilesArtifactMountpoints() {
		artifactUrl, found := networkCtx.filesArtifactUrls[filesArtifactId]
		if !found {
			return nil, nil, stacktrace.Propagate(
				err,
				"Service requested file artifact '%v', but the network" +
					"context doesn't have a URL for that file artifact; this is a bug with Kurtosis itself",
				filesArtifactId)
		}
		artifactUrlToMountDirpath[string(artifactUrl)] = mountDirpath
	}
	logrus.Tracef("Successfully created files artifact URL -> mount dirpaths map")

	logrus.Tracef("Creating start command for service...")
	serviceIpAddr := registerServiceResp.IpAddr
	startCmdArgs, err := initializer.GetStartCommand(generatedFilesAbsoluteFilepathsOnService, serviceIpAddr)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to create start command")
	}
	logrus.Tracef("Successfully created start command for service")

	logrus.Tracef("Starting new service with Kurtosis API...")
	startServiceArgs := &core_api_bindings.StartServiceArgs{
		ServiceId:                   string(serviceId),
		DockerImage:                 initializer.GetDockerImage(),
		UsedPorts:                   initializer.GetUsedPorts(),
		StartCmdArgs:                startCmdArgs,
		DockerEnvVars:               map[string]string{}, // TODO actually support Docker env vars!
		SuiteExecutionVolMntDirpath: initializer.GetTestVolumeMountpoint(),
		FilesArtifactMountDirpaths:  artifactUrlToMountDirpath,
	}
	if _, err := networkCtx.client.StartService(ctx, startServiceArgs); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the service with the Kurtosis API")
	}
	logrus.Tracef("Successfully started service with Kurtosis API")

	logrus.Tracef("Creating service interface...")
	service := initializer.GetService(serviceId, serviceIpAddr)
	logrus.Tracef("Successfully created service interface")

	networkCtx.services[serviceId] = service

	availabilityChecker := services.NewDefaultAvailabilityChecker(serviceId, service)

	return service, availabilityChecker, nil
}
 */
    }

    pub fn get_service(&self) {
/*
func (networkCtx *NetworkContext) GetService(serviceId services.ServiceID) (services.Service, error) {
	networkCtx.mutex.Lock()
	defer networkCtx.mutex.Unlock()

	service, found := networkCtx.services[serviceId]
	if !found {
		return nil, stacktrace.NewError("No service found with ID '%v'", serviceId)
	}

	return service, nil
}
 */
    }

    pub fn remove_service(&self) {
/*
func (networkCtx *NetworkContext) RemoveService(serviceId services.ServiceID, containerStopTimeoutSeconds uint64) error {
	networkCtx.mutex.Lock()
	defer networkCtx.mutex.Unlock()

	logrus.Debugf("Removing service '%v'...", serviceId)
	args := &core_api_bindings.RemoveServiceArgs{
		ServiceId:                   string(serviceId),
		// NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
		//  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
		// Independent of adding/removing them from the network
		ContainerStopTimeoutSeconds: containerStopTimeoutSeconds,
	}
	if _, err := networkCtx.client.RemoveService(context.Background(), args); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v' from the network", serviceId)
	}
	delete(networkCtx.services, serviceId)
	logrus.Debugf("Successfully removed service ID %v", serviceId)
	return nil
}
 */
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
}