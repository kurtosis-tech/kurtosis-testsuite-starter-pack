/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis-go/lib/client/artifact_id_provider"
	"github.com/kurtosis-tech/kurtosis-go/lib/core_api/bindings"
	"github.com/kurtosis-tech/kurtosis-go/lib/kurtosis_service"
	"github.com/kurtosis-tech/kurtosis-go/lib/kurtosis_service/method_types"
	"github.com/kurtosis-tech/kurtosis-go/lib/services"
	"github.com/kurtosis-tech/kurtosis-go/lib/test_suite_docker_consts/test_suite_container_mountpoints"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

const (
	// NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
	//  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
	// Independent of adding/removing them from the network
	removeServiceContainerStopTimeout = 10 * time.Second

	// This will alwyas resolve to the default partition ID (regardless of whether such a partition exists in the network,
	//  or it was repartitioned away)
	defaultPartitionId PartitionID = ""
)

type NetworkContext struct {
	client bindings.TestExecutionServiceClient

	filesArtifactUrls map[services.FilesArtifactID]string
}


/*
Creates a new NetworkContext object with the given parameters.

Args:
	client: The Kurtosis API client that the NetworkContext will use for modifying the state of the testnet
	filesArtifactUrls: The mapping of filesArtifactId -> URL for the artifacts that the testsuite will use
*/
func NewNetworkContext(
		client bindings.TestExecutionServiceClient,
		filesArtifactUrls map[services.FilesArtifactID]string) *NetworkContext {
	return &NetworkContext{
		client: client,
		filesArtifactUrls: filesArtifactUrls,
	}
}

/*
Adds a service to the network in the default partition with the given service ID

NOTE: If the network has been repartitioned and the default partition hasn't been preserved, you should use
	AddServiceToPartition instead.

Args:
	serviceId: The service ID that will be used to identify this node in the network.
	initializer: The Docker container initializer that contains the logic for starting the service

Return:
	service: The new service
*/
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

/*
Adds a service to the network with the given service ID, created using the given configuration ID.

NOTE: If the network hasn't been repartitioned yet, the PartitionID should be an empty string to add to the default
	partition.

Args:
	serviceId: The service ID that will be used to identify this node in the network.
	partitionId: The partition ID to add the service to
	initializer: The Docker container initializer that contains the logic for starting the service

Return:
	service.Service: The new service
*/
func (networkCtx *NetworkContext) AddServiceToPartition(
		serviceId services.ServiceID,
		partitionId PartitionID,
		initializer services.DockerContainerInitializer) (services.Service, services.AvailabilityChecker, error) {
	ctx := context.Background()

	logrus.Tracef("Registering new service ID with Kurtosis API...")
	registerServiceArgs := &bindings.RegisterServiceArgs{
		ServiceId:       string(serviceId),
		PartitionId:     string(partitionId),
		FilesToGenerate: initializer.GetFilesToMount(),
	}
	registerServiceResp, err := networkCtx.client.RegisterService(ctx, registerServiceArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering service with ID '%v' with the Kurtosis API")
	}
	logrus.Tracef("New service successfully registered with Kurtosis API")

	generatedFilesRelativeFilepaths := registerServiceResp.GeneratedFilesRelativeFilepaths
	generatedFilesFps := map[string]*os.File{}
	generatedFilesAbsoluteFilepaths := map[string]string{}
	for fileId, relativeFilepath := range generatedFilesRelativeFilepaths {
		absoluteFilepath := path.Join(test_suite_container_mountpoints.SuiteExVolMountpoint, relativeFilepath)
		generatedFilesAbsoluteFilepaths[fileId] = absoluteFilepath
		fp, err := os.Create(absoluteFilepath)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"Could not open generated file '%v' for writing",
				fileId)
		}
		defer fp.Close()
		generatedFilesFps[fileId] = fp
	}

	logrus.Trace("Initializing generated files...")
	if err := initializer.InitializeMountedFiles(generatedFilesFps); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred initializing the generated files")
	}
	logrus.Trace("Successfully initialized generated files")


	logrus.Tracef("Creating files artifact mount dirpaths map...")
	filesArtifactMountDirpaths := map[string]string{}
	for filesArtifactId, mountDirpath := range initializer.GetFilesArtifactMountpoints() {
		filesArtifactMountDirpaths[string(filesArtifactId)] = mountDirpath
	}
	logrus.Tracef("Successfully created files artifact mount dirpaths map")

	logrus.Tracef("Creating start command for service...")
	serviceIpAddr := registerServiceResp.IpAddr
	startCmdArgs, err := initializer.GetStartCommand(generatedFilesAbsoluteFilepaths, serviceIpAddr)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to create start command")
	}
	logrus.Tracef("Successfully created start command for service")

	logrus.Tracef("Starting new service with Kurtosis API...")
	dockerImage := initializer.GetDockerImage()
	startServiceArgs := &bindings.StartServiceArgs{
		ServiceId:                   string(serviceId),
		DockerImage:                 initializer.GetDockerImage(),
		UsedPorts:                   initializer.GetUsedPorts(),
		StartCmdArgs:                startCmdArgs,
		DockerEnvVars:               map[string]string{}, // TODO actually support Docker env vars!
		SuiteExecutionVolMntDirpath: initializer.GetTestVolumeMountpoint(),
		FilesArtifactMountDirpaths:  filesArtifactMountDirpaths,
	}
	if _, err := networkCtx.client.StartService(ctx, startServiceArgs); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the service with the Kurtosis API")
	}
	logrus.Tracef("Successfully started service with Kurtosis API")

	logrus.Tracef("Getting service from IP...")
	service := initializer.GetService(serviceIpAddr)
	logrus.Tracef("Successfully got service from IP")

	availabilityChecker := services.NewDefaultAvailabilityChecker(serviceId, service)

	return service, availabilityChecker, nil
}

/*
Stops the container with the given service ID, and removes it from the network.
*/
func (networkCtx *NetworkContext) RemoveService(serviceId services.ServiceID, containerStopTimeoutSeconds int) error {
	logrus.Debugf("Removing service '%v'...", serviceId)
	args := &bindings.RemoveServiceArgs{
		ServiceId:                   string(serviceId),
		ContainerStopTimeoutSeconds: removeServiceContainerStopTimeout,
	}
	if _, err := networkCtx.client.RemoveService(context.Background(), args); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v' from the network", serviceId)
	}
	logrus.Debugf("Successfully removed service ID %v", serviceId)
	return nil
}

/*
Constructs a new repartitioner builder in preparation for a repartition.

Args:
	isDefaultPartitionConnectionBlocked: If true, when the connection details between two partitions aren't specified
		during a repartition then traffic between them will be blocked by default
 */
func (networkCtx NetworkContext) GetRepartitionerBuilder(isDefaultPartitionConnectionBlocked bool) *RepartitionerBuilder {
	// This function doesn't need a mutex lock because (as of 2020-12-28) it doesn't touch internal state whatsoever
	return newRepartitionerBuilder(isDefaultPartitionConnectionBlocked)
}

/*
Repartitions the network using the given repartitioner. A repartitioner builder can be constructed using the
	NewRepartitionerBuilder method of this network context object.
 */
func (networkCtx *NetworkContext) RepartitionNetwork(repartitioner *Repartitioner) error {
	partitionServices := map[string]*bindings.PartitionServices{}
	for partitionId, serviceIdSet := range repartitioner.partitionServices {
		serviceIdStrPseudoSet := map[string]bool{}
		for _, serviceId := range serviceIdSet.getElems() {
			serviceIdStr := string(serviceId)
			serviceIdStrPseudoSet[serviceIdStr] = true
		}
		partitionIdStr := string(partitionId)
		partitionServices[partitionIdStr] = serviceIdStrPseudoSet
	}

	serializablePartConns := map[string]map[string]method_types.SerializablePartitionConnection{}
	for partitionAId, partitionAConns := range repartitioner.partitionConnections {
		serializablePartAConns := map[string]method_types.SerializablePartitionConnection{}
		for partitionBId, unserializableConn := range partitionAConns {
			partitionBIdStr := string(partitionBId)
			serializableConn := makePartConnSerializable(unserializableConn)
			serializablePartAConns[partitionBIdStr] = serializableConn
		}
		partitionAIdStr := string(partitionAId)
		serializablePartConns[partitionAIdStr] = serializablePartAConns
	}

	serializableDefaultConn := makePartConnSerializable(repartitioner.defaultConnection)
	
	repartitionArgs := &bindings.RepartitionArgs{
		PartitionServices:    partitionServices,
		PartitionConnections: serializablePartConns,
		DefaultConnection:    serializableDefaultConn,
	}

	if err := networkCtx.kurtosisService.Repartition(partitionServices, serializablePartConns, serializableDefaultConn); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return nil
}

// ============================================================================================
//                                    Private helper methods
// ============================================================================================
func makePartConnSerializable(connection PartitionConnection) method_types.SerializablePartitionConnection {
	return method_types.SerializablePartitionConnection{
		IsBlocked: connection.IsBlocked,
	}
}
