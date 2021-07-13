/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package network_partition_test

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/example-microservice/api/api_service_client"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	defaultPartitionId   networks.PartitionID = ""
	apiPartitionId       networks.PartitionID = "api"
	datastorePartitionId networks.PartitionID = "datastore"
	datastorePort                             = 1323
	apiServicePort                            = 2434
	datastoreServiceId   services.ServiceID   = "datastore"
	api1ServiceId        services.ServiceID   = "api1"
	api2ServiceId        services.ServiceID   = "api2"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxNumPolls       = 15

	testPersonId = 46
)

type NetworkPartitionTest struct {
	datstoreImage string
	apiImage      string
}

func NewNetworkPartitionTest(datstoreImage string, apiImage string) *NetworkPartitionTest {
	return &NetworkPartitionTest{datstoreImage: datstoreImage, apiImage: apiImage}
}

func (test NetworkPartitionTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(
		60,
	).WithRunTimeoutSeconds(
		60,
	).WithPartitioningEnabled(true)
}

// Instantiates the network with no partition and one person in the datatstore
func (test NetworkPartitionTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		"kurtosistech/example-microservices_datastore",
	).WithUsedPorts(
		map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
	).Build()

	generateRunConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		return services.NewContainerRunConfigBuilder().Build(), nil
	}

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, containerCreationConfig, generateRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)

	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceContext.GetIPAddress(), datastorePort)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	apiClient, err := test.addApiService(networkCtx, api1ServiceId, defaultPartitionId, datastoreClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v'", api1ServiceId)
	}

	if err := apiClient.AddPerson(testPersonId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the test person in preparation for the test")
	}
	if err := apiClient.IncrementBooksRead(testPersonId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred test person's books read in preparation for the test")
	}

	return networkCtx, nil
}

func (test NetworkPartitionTest) Run(network networks.Network) error {
	// Go doesn't have generics so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	logrus.Info("Partitioning API and datastore services off from each other...")
	if err := repartitionNetwork(castedNetwork, true, false); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the network to block access between API <-> datastore")
	}
	logrus.Info("Repartition complete")

	datastoreServiceContext, err := castedNetwork.GetServiceContext(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service context")
	}
	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceContext.GetIPAddress(), datastorePort)

	logrus.Info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")
	apiServiceContext, err := castedNetwork.GetServiceContext(api1ServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API 1 service context")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), apiServicePort)
	if err := apiClient.IncrementBooksRead(testPersonId); err == nil {
		return stacktrace.NewError("Expected the book increment call via API 1 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown")
	} else {
		logrus.Infof("Incrementing books read via API 1 threw the following error as expected due to network partition: %v", err)
	}

	// Adding another API service while the partition is in place ensures that partitiong works even when you add a node
	logrus.Info("Adding second API container, to ensure adding a network under partition works...")

	apiClient2, err := test.addApiService(
		castedNetwork,
		api2ServiceId,
		apiPartitionId,
		datastoreClient,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the second API service to the network")
	}
	logrus.Info("Second API container added successfully")

	logrus.Info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")
	if err := apiClient2.IncrementBooksRead(testPersonId); err == nil {
		return stacktrace.NewError("Expected the book increment call via API 2 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown")
	} else {
		logrus.Infof("Incrementing books read via API 2 threw the following error as expected due to network partition: %v", err)
	}

	// Now, open the network back up
	logrus.Info("Repartitioning to heal partition between API and datastore...")
	if err := repartitionNetwork(castedNetwork, false, true); err != nil {
		return stacktrace.Propagate(err, "An error occurred healing the partition")
	}
	logrus.Info("Partition healed successfully")

	logrus.Info("Making another call via API 1 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	if err := apiClient.IncrementBooksRead(testPersonId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 1, even though the partition should have been "+
				"healed by the goroutine",
		)
	}
	logrus.Info("Successfully incremented books read via API 1, indicating that the partition has healed successfully!")

	logrus.Info("Making another call via API 2 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	if err := apiClient2.IncrementBooksRead(testPersonId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 2, even though the partition should have been "+
				"healed by the goroutine",
		)
	}
	logrus.Info("Successfully incremented books read via API 2, indicating that the partition has healed successfully!")
	return nil
}

// ========================================================================================================
//                                     Private helper functions
// ========================================================================================================
func (test NetworkPartitionTest) addApiService(
	networkCtx *networks.NetworkContext,
	serviceId services.ServiceID,
	partitionId networks.PartitionID,
	datastoreServiceClient *datastore_service_client.DatastoreClient) (*api_service_client.APIClient, error) {

	configFileKey := "config-file"

	type config struct {
		DatastoreIp   string `json:"datastoreIp"`
		DatastorePort int    `json:"datastorePort"`
	}

	configInitializingFunc := func(fp *os.File) error {
		logrus.Debugf("Datastore IP: %v , port: %v", datastoreServiceClient.IpAddr(), datastoreServiceClient.Port())
		configObj := config{
			DatastoreIp:   datastoreServiceClient.IpAddr(),
			DatastorePort: datastoreServiceClient.Port(),
		}
		configBytes, err := json.Marshal(configObj)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
		}

		logrus.Debugf("API config JSON: %v", string(configBytes))

		if _, err := fp.Write(configBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
		}

		return nil
	}

	apiServiceContainerCreationConfig := services.NewContainerCreationConfigBuilder(
		"kurtosistech/example-microservices_api",
	).WithUsedPorts(
		map[string]bool{fmt.Sprintf("%v/tcp", apiServicePort): true},
	).WithGeneratedFiles(map[string]func(*os.File) error{
		configFileKey: configInitializingFunc,
	}).Build()

	apiServiceGenerateRunConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		configFilepath, found := generatedFileFilepaths[configFileKey]
		if !found {
			return nil, stacktrace.NewError("No filepath found for config file key '%v'", configFileKey)
		}
		startCmd := []string{
			"./api.bin",
			"--config",
			configFilepath,
		}
		result := services.NewContainerRunConfigBuilder().WithCmdOverride(startCmd).Build()
		return result, nil
	}

	apiServiceContext, hostPortBindings, err := networkCtx.AddServiceToPartition(serviceId, partitionId, apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), apiServicePort)
	err = apiClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
	}

	logrus.Infof("Added API service '%v' with host port bindings: %+v", serviceId, hostPortBindings)
	return apiClient, nil
}

/*
Creates a repartitioner that will partition the network between the API & datastore services, with the connection between them configurable
*/
func repartitionNetwork(
	networkCtx *networks.NetworkContext,
	isConnectionBlocked bool,
	isApi2ServiceAddedYet bool) error {
	apiPartitionServiceIds := map[services.ServiceID]bool{
		api1ServiceId: true,
	}
	if isApi2ServiceAddedYet {
		apiPartitionServiceIds[api2ServiceId] = true
	}

	partitionServices := map[networks.PartitionID]map[services.ServiceID]bool{
		apiPartitionId: apiPartitionServiceIds,
		datastorePartitionId: {
			datastoreServiceId: true,
		},
	}
	partitionConnections := map[networks.PartitionID]map[networks.PartitionID]*core_api_bindings.PartitionConnectionInfo{
		apiPartitionId: {
			datastorePartitionId: &core_api_bindings.PartitionConnectionInfo{
				IsBlocked: isConnectionBlocked,
			},
		},
	}
	defaultPartitionConnection := &core_api_bindings.PartitionConnectionInfo{
		IsBlocked: false,
	}
	if err := networkCtx.RepartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred repartitioning the network with isConnectionBlocked = %v",
			isConnectionBlocked)
	}
	return nil
}
