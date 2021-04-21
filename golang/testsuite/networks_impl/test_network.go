/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks_impl

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
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

type TestNetwork struct {
	networkCtx            *networks.NetworkContext
	datastoreServiceImage string
	apiServiceImage       string
	datastoreService      *datastore.DatastoreService
	apiServices           map[services.ServiceID]*api.ApiService
	nextApiServiceId      int
}

func NewTestNetwork(networkCtx *networks.NetworkContext, datastoreServiceImage string, apiServiceImage string) *TestNetwork {
	return &TestNetwork{
		networkCtx:            networkCtx,
		datastoreServiceImage: datastoreServiceImage,
		apiServiceImage:       apiServiceImage,
		datastoreService:      nil,
		apiServices:           map[services.ServiceID]*api.ApiService{},
		nextApiServiceId:      0,
	}
}

func (network *TestNetwork) AddDatastore() error {
	if (network.datastoreService != nil) {
		return stacktrace.NewError("Cannot add datastore service to network; datastore already exists!")
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
	return nil
}

func (network *TestNetwork) GetDatastore() *datastore.DatastoreService {
	return network.datastoreService
}

func (network *TestNetwork) AddApiService() (services.ServiceID, error) {
	if (network.datastoreService == nil) {
		return "", stacktrace.NewError("Cannot add API service to network; no datastore service exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	configFactory := api.NewApiContainerConfigFactory(network.apiServiceImage, network.datastoreService)
	uncastedApiService, hostPortBindings, checker, err := network.networkCtx.AddService(serviceId, configFactory)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred adding the API service")
	}
	if err := checker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred waiting for the API service to start")
	}
	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	castedApiService := uncastedApiService.(*api.ApiService)
	network.apiServices[serviceId] = castedApiService
	return serviceId, nil
}

func (network *TestNetwork) GetApiService(serviceId services.ServiceID) (*api.ApiService, error) {
	service, found := network.apiServices[serviceId]
	if !found {
		return nil, stacktrace.NewError("No API service with ID '%v' has been added", serviceId)
	}
	return service, nil
}
