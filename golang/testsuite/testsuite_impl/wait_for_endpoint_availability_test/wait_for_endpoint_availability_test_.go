/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package wait_for_endpoint_availability_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)
const (
	datastoreServiceId services.ServiceID = "datastore"

	healthCheckUrlSlug = "health"
	healthyValue       = "healthy"

	waitForStartupTimeBetweenPolls = 1
	waitForStartupMaxPolls = 15
	waitInitialDelaySeconds = 1
)


type WaitForEndpointAvailabilityTest struct {
	datastoreImage string
}

func NewWaitForEndpointAvailabilityTest(datastoreImage string) *WaitForEndpointAvailabilityTest {
	return &WaitForEndpointAvailabilityTest{datastoreImage: datastoreImage}
}

func (test WaitForEndpointAvailabilityTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test WaitForEndpointAvailabilityTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	return networkCtx, nil
}

func (test WaitForEndpointAvailabilityTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetworkContext := network.(*networks.NetworkContext)

	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	_, _,  err := castedNetworkContext.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	port := uint32(datastore.Port)

	if err := castedNetworkContext.WaitForEndpointAvailability(datastoreServiceId, port, healthCheckUrlSlug, waitInitialDelaySeconds, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Service: %v is available", datastoreServiceId)

	return nil
}
