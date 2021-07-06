/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_test

import (
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	datastoreServiceId services.ServiceID = "datastore"

	testKey = "test-key"
	testValue = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls = 15
)

type BasicDatastoreTest struct {
	datastoreImage string
}

func NewBasicDatastoreTest(datastoreImage string) *BasicDatastoreTest {
	return &BasicDatastoreTest{datastoreImage: datastoreImage}
}

func (test BasicDatastoreTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test BasicDatastoreTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	serviceContext, hostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(serviceContext.GetIPAddress(), datastore.Port)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test BasicDatastoreTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	serviceInfo, err := castedNetwork.GetServiceInfo(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service info")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(serviceInfo.GetIPAddress().String(), datastore.Port)

	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	exists, err := datastoreClient.Exists(testKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if the test key exists")
	}
	if exists {
		return stacktrace.NewError("Test key should not exist yet")
	}
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	if err := datastoreClient.Upsert(testKey, testValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred upserting the test key")
	}
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	value, err := datastoreClient.Get(testKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key after upload")
	}
	if value != testValue {
		return stacktrace.NewError("Returned value '%v' != test value '%v'", value, testValue)
	}
	logrus.Info("Value verified")
	return nil
}