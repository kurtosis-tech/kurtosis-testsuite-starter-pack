/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls = 15

	testKey = "test-key"
	testValue = "test-value"
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
	_, hostPortBindings, availabilityChecker, err := networkCtx.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}
	if err := availabilityChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test BasicDatastoreTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	uncastedService, err := castedNetwork.GetService(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service")
	}

	// Necessary again due to no Go generics
	castedService := uncastedService.(*datastore.DatastoreService)

	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	exists, err := castedService.Exists(testKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if the test key exists")
	}
	if exists {
		return stacktrace.NewError("Test key should not exist yet")
	}
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	if err := castedService.Upsert(testKey, testValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred upserting the test key")
	}
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	value, err := castedService.Get(testKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key after upload")
	}
	if value != testValue {
		return stacktrace.NewError("Returned value '%v' != test value '%v'", value, testValue)
	}
	logrus.Info("Value verified")
	return nil
}