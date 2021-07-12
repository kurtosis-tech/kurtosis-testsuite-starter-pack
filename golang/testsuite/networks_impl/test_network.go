/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks_impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/example-microservice/api/api_service_client"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	apiServiceIdPrefix                    = "api-"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxNumPolls       = 15
)

//  A custom Network implementation is intended to make test-writing easier by wrapping low-level
//    NetworkContext calls with custom higher-level business logic
type TestNetwork struct {
	networkCtx                *networks.NetworkContext
	datastoreServiceImage     string
	apiServiceImage           string
	datastoreClient           *datastore_service_client.DatastoreClient
	personModifyingApiClient  *api_service_client.APIClient
	personRetrievingApiClient *api_service_client.APIClient
	nextApiServiceId          int
}

func NewTestNetwork(networkCtx *networks.NetworkContext, datastoreServiceImage string, apiServiceImage string) *TestNetwork {
	return &TestNetwork{
		networkCtx:                networkCtx,
		datastoreServiceImage:     datastoreServiceImage,
		apiServiceImage:           apiServiceImage,
		datastoreClient:           nil,
		personModifyingApiClient:  nil,
		personRetrievingApiClient: nil,
		nextApiServiceId:          0,
	}
}

//  Custom network implementations usually have a "setup" method (possibly parameterized) that is used
//   in the Test.Setup function of each test
func (network *TestNetwork) SetupDatastoreAndTwoApis() error {

	if network.datastoreClient != nil {
		return stacktrace.NewError("Cannot add datastore client to network; datastore client already exists!")
	}

	if network.personModifyingApiClient != nil || network.personRetrievingApiClient != nil {
		return stacktrace.NewError("Cannot add API services to network; one or more API services already exists")
	}

	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		"kurtosistech/example-microservices_datastore",
		"/test-volume",
	).WithUsedPorts(
		map[string]bool{"1323/tcp": true},
	).Build()

	generateRunConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		return services.NewContainerRunConfigBuilder().Build(), nil
	}

	datastoreServiceContext, hostPortBindings, err := network.networkCtx.AddService(datastoreServiceId, containerCreationConfig, generateRunConfigFunc)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceContext.GetIPAddress(), datastore.Port)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)

	network.datastoreClient = datastoreClient

	personModifyingApiClient, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-modifying API client")
	}
	network.personModifyingApiClient = personModifyingApiClient

	personRetrievingApiClient, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-retrieving API client")
	}
	network.personRetrievingApiClient = personRetrievingApiClient

	return nil
}

//  Custom network implementations will also usually have getters, to retrieve information about the
//   services created during setup
func (network *TestNetwork) GetPersonModifyingApiClient() (*api_service_client.APIClient, error) {
	if network.personModifyingApiClient == nil {
		return nil, stacktrace.NewError("No person-modifying API client exists")
	}
	return network.personModifyingApiClient, nil
}
func (network *TestNetwork) GetPersonRetrievingApiClient() (*api_service_client.APIClient, error) {
	if network.personRetrievingApiClient == nil {
		return nil, stacktrace.NewError("No person-retrieving API client exists")
	}
	return network.personRetrievingApiClient, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (network *TestNetwork) addApiService() (*api_service_client.APIClient, error) {

	if network.datastoreClient == nil {
		return nil, stacktrace.NewError("Cannot add API service to network; no datastore client exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	configFileKey := "config-file"

	type config struct {
		DatastoreIp   string `json:"datastoreIp"`
		DatastorePort int    `json:"datastorePort"`
	}

	configInitializingFunc := func(fp *os.File) error {
		logrus.Debugf("Datastore IP: %v , port: %v", network.datastoreClient.IpAddr(), network.datastoreClient.Port())
		configObj := config{
			DatastoreIp:   network.datastoreClient.IpAddr(),
			DatastorePort: network.datastoreClient.Port(),
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
		"/test-volume",
	).WithUsedPorts(
		map[string]bool{"2434/tcp": true},
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

	apiServiceContext, hostPortBindings, err := network.networkCtx.AddService(serviceId, apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), 2434)

	err = apiClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	return apiClient, nil
}
