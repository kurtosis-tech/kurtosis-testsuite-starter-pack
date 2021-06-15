/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks_impl

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/api"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	apiServiceIdPrefix = "api-"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxNumPolls = 15
)

//  A custom Network implementation is intended to make test-writing easier by wrapping low-level
//    NetworkContext calls with custom higher-level business logic
type TestNetwork struct {
	networkCtx            *networks.NetworkContext
	datastoreServiceImage string
	apiServiceImage       string
	datastoreService      *datastore.DatastoreService
	personModifyingApiService *api.ApiService
	personRetrievingApiService *api.ApiService
	nextApiServiceId      int
}

func NewTestNetwork(networkCtx *networks.NetworkContext, datastoreServiceImage string, apiServiceImage string) *TestNetwork {
	return &TestNetwork{
		networkCtx:            networkCtx,
		datastoreServiceImage: datastoreServiceImage,
		apiServiceImage:       apiServiceImage,
		datastoreService:      nil,
		personModifyingApiService: nil,
		personRetrievingApiService: nil,
		nextApiServiceId:      0,
	}
}

//  Custom network implementations usually have a "setup" method (possibly parameterized) that is used
//   in the Test.Setup function of each test
func (network *TestNetwork) SetupDatastoreAndTwoApis() error {
	if network.datastoreService != nil {
		return stacktrace.NewError("Cannot add datastore service to network; datastore already exists!")
	}
	if network.personModifyingApiService != nil || network.personRetrievingApiService != nil {
		return stacktrace.NewError("Cannot add API services to network; one or more API services already exists")
	}

	configFactory := datastore.NewDatastoreContainerConfigFactory(network.datastoreServiceImage)
	uncastedDatastore, hostPortBindings, checker, err := network.networkCtx.AddService(datastoreServiceId, configFactory)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}
	if err := checker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to start")
	}
	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	castedDatastore := uncastedDatastore.(*datastore.DatastoreService)
	network.datastoreService = castedDatastore

	personModifyingApiService, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-modifying API service")
	}
	network.personModifyingApiService = personModifyingApiService

	personRetrievingApiServiceId, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-retrieving API service")
	}
	network.personRetrievingApiService = personRetrievingApiServiceId

	return nil
}

//  Custom network implementations will also usually have getters, to retrieve information about the
//   services created during setup
func (network TestNetwork) GetPersonModifyingApiService() (*api.ApiService, error) {
	if network.personModifyingApiService == nil {
		return nil, stacktrace.NewError("No person-modifying API service exists")
	}
	return network.personModifyingApiService, nil
}
func (network TestNetwork) GetPersonRetrievingApiService() (*api.ApiService, error) {
	if network.personRetrievingApiService == nil {
		return nil, stacktrace.NewError("No person-retrieving API service exists")
	}
	return network.personRetrievingApiService, nil
}


// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (network *TestNetwork) addApiService() (*api.ApiService, error) {
	if network.datastoreService == nil {
		return nil, stacktrace.NewError("Cannot add API service to network; no datastore service exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	configFactory := api.NewApiContainerConfigFactory(network.apiServiceImage, network.datastoreService)
	uncastedApiService, hostPortBindings, checker, err := network.networkCtx.AddService(serviceId, configFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}
	if err := checker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API service to start")
	}
	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	castedApiService := uncastedApiService.(*api.ApiService)
	return castedApiService, nil
}

