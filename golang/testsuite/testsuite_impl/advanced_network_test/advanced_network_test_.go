/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package advanced_network_test

import (
	"context"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/golang/testsuite/networks_impl"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	testPersonId = "46"
)

type AdvancedNetworkTest struct {
	datastoreServiceImage string
	apiServiceImage string
}

func NewAdvancedNetworkTest(datastoreServiceImage string, apiServiceImage string) *AdvancedNetworkTest {
	return &AdvancedNetworkTest{datastoreServiceImage: datastoreServiceImage, apiServiceImage: apiServiceImage}
}

func (test *AdvancedNetworkTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test *AdvancedNetworkTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	network := networks_impl.NewTestNetwork(networkCtx, test.datastoreServiceImage, test.apiServiceImage)
	// Note how setup logic has been pushed into a custom Network implementation, to make test-writing easy
	if err := network.SetupDatastoreAndTwoApis(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred setting up the network")
	}
	return network, nil
}

func (test *AdvancedNetworkTest) Run(network networks.Network) error {
	ctx := context.Background()

	castedNetwork := network.(*networks_impl.TestNetwork)
	personModifierClient, personModifyingApiClientCloseFunc, err := castedNetwork.GetPersonModifyingApiClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the person-modifying API client")
	}
	defer func() {
		if err := personModifyingApiClientCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the person modifying API client, but doing so threw an error:\n%v", err)
		}
	}()

	personRetrieverClient, personRetrieverApiClientCloseFunc, err := castedNetwork.GetPersonRetrievingApiClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the person-retrieving API client")
	}
	defer func() {
		if err := personRetrieverApiClientCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the person modifying API client, but doing so threw an error:\n%v", err)
		}
	}()

	logrus.Infof("Adding test person via person-modifying API client...")
	addPersonArgs := &example_api_server_rpc_api_bindings.AddPersonArgs{
		PersonId: testPersonId,
	}
	if _, err := personModifierClient.AddPerson(ctx, addPersonArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding test person with ID '%v'", testPersonId)
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read through person-modifying API client...")
	incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
		PersonId: testPersonId,
	}
	if _, err := personModifierClient.IncrementBooksRead(ctx, incrementBooksReadArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred incrementing the number of books read")
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read person-retrieving API client...")
	getPersonArgs := &example_api_server_rpc_api_bindings.GetPersonArgs{
		PersonId: testPersonId,
	}
	getPersonResponse, err := personRetrieverClient.GetPerson(ctx, getPersonArgs)
	if err != nil {
		return stacktrace.NewError("An error occurred getting the test person with ID '%v'", testPersonId)
	}
	logrus.Info("Retrieved test person")

	personBooksRead := getPersonResponse.GetBooksRead()

	if personBooksRead != 1 {
		return stacktrace.NewError(
			"Expected number of books read to be incremented, but was '%v'",
			personBooksRead,
		)
	}

	return nil
}
