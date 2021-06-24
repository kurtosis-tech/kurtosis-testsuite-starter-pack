/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package check_service_availability_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
)
const (
	datastoreServiceId services.ServiceID = "datastore"

	healthCheckUrlSlug = "health"
	healthyValue       = "healthy"

	waitForStartupTimeBetweenPolls = 1
	waitForStartupMaxPolls = 15
)


type CheckServiceAvailabilityTest struct {
	datastoreImage string
}

func NewCheckServiceAvailabilityTest(datastoreImage string) *CheckServiceAvailabilityTest {
	return &CheckServiceAvailabilityTest{datastoreImage: datastoreImage}
}

func (test CheckServiceAvailabilityTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test CheckServiceAvailabilityTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	_, hostPortBindings, _, err := networkCtx.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test CheckServiceAvailabilityTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	//TODO i'm not pretty sure if this should be the right place to get the port
	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	port := strconv.Itoa(datastoreConfigFactory.GetPort())

	if err := castedNetwork.CheckServiceAvailability(datastoreServiceId, port, healthCheckUrlSlug, 1, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Service: %v is available", datastoreServiceId)

	return nil
}