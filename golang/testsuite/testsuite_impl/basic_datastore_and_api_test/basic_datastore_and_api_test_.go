/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_and_api_test

import (
	"github.com/kurtosis-tech/example-microservice/api/api_service_client"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/api"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	apiServiceId services.ServiceID = "api"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls = 15

	testPersonId = 23
	testNumBooksRead = 3
)

type BasicDatastoreAndApiTest struct {
	datstoreImage string
	apiImage string
}

func NewBasicDatastoreAndApiTest(datstoreImage string, apiImage string) *BasicDatastoreAndApiTest {
	return &BasicDatastoreAndApiTest{datstoreImage: datstoreImage, apiImage: apiImage}
}

func (b BasicDatastoreAndApiTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (b BasicDatastoreAndApiTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(b.datstoreImage)
	datastoreServiceInfo, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceInfo.GetIPAddress().String(), datastore.Port)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)

	apiConfigFactory := api.NewApiContainerConfigFactory(b.apiImage, datastoreServiceInfo.GetIPAddress().String(), datastore.Port)
	apiServiceInfo, apiSvcHostPortBindings, err := networkCtx.AddService(apiServiceId, apiConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceInfo.GetIPAddress().String(), api.Port)

	err = apiClient.WaitForHealthy(waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", apiSvcHostPortBindings)
	return networkCtx, nil
}


func (b BasicDatastoreAndApiTest) Run(network networks.Network) error {
	// Go doesn't have generics so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	serviceInfo, err := castedNetwork.GetServiceInfo(apiServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API service info")
	}

	apiClient := api_service_client.NewAPIClient(serviceInfo.GetIPAddress().String(), api.Port)

	logrus.Infof("Verifying that person with test ID '%v' doesn't already exist...", testPersonId)
	if _, err = apiClient.GetPerson(testPersonId); err == nil {
		return stacktrace.NewError("Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
	}
	logrus.Infof("Verified that test person doesn't already exist")

	logrus.Infof("Adding test person with ID '%v'...", testPersonId)
	if err := apiClient.AddPerson(testPersonId); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding person with test ID '%v'", testPersonId)
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read by %v...", testNumBooksRead)
	for i := 0; i < testNumBooksRead; i++ {
		if err := apiClient.IncrementBooksRead(testPersonId); err != nil {
			return stacktrace.Propagate(err, "An error occurred incrementing the number of books read")
		}
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read...")
	person, err := apiClient.GetPerson(testPersonId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test person to verify the number of books read")
	}
	logrus.Info("Retrieved test person")

	if person.BooksRead != testNumBooksRead {
		return stacktrace.NewError(
			"Expected number of book read '%v' != actual number of books read '%v'",
			testNumBooksRead,
			person.BooksRead,
		)
	}

	return nil
}