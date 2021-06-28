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


type WaitForEndpointAbailabilityTest struct {
	datastoreImage string
}

func NewWaitForEnpointAvailabilityTest(datastoreImage string) *WaitForEndpointAbailabilityTest {
	return &WaitForEndpointAbailabilityTest{datastoreImage: datastoreImage}
}

func (test WaitForEndpointAbailabilityTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test WaitForEndpointAbailabilityTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	_, hostPortBindings, _, err := networkCtx.AddService(datastoreServiceId, datastoreConfigFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test WaitForEndpointAbailabilityTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	datastoreConfigFactory := datastore.NewDatastoreContainerConfigFactory(test.datastoreImage)
	port := uint32(datastoreConfigFactory.GetPort())

	if err := castedNetwork.WaitForEndpointAvailability(datastoreServiceId, port, healthCheckUrlSlug, waitInitialDelaySeconds, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Service: %v is available", datastoreServiceId)

	return nil
}