/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package network_partition_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/api"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	defaultPartitionId networks.PartitionID = ""
	apiPartitionId networks.PartitionID = "api"
	datastorePartitionId networks.PartitionID = "datastore"

	datastoreServiceId services.ServiceID = "datastore"
	api1ServiceId      services.ServiceID = "api1"
	api2ServiceId      services.ServiceID = "api2"


	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxNumPolls = 30

	testPersonId = 46
)

type NetworkPartitionTest struct {
	datstoreImage string
	apiImage string
}

func NewNetworkPartitionTest(datstoreImage string, apiImage string) *NetworkPartitionTest {
	return &NetworkPartitionTest{datstoreImage: datstoreImage, apiImage: apiImage}
}

// Instantiates the network with no partition and one person in the datatstore
func (test NetworkPartitionTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreInitializer := datastore.NewDatastoreContainerInitializer(test.datstoreImage)
	uncastedDatastoreSvc, datastoreChecker, err := networkCtx.AddService(datastoreServiceId, datastoreInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}
	if err := datastoreChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to start")
	}

	// Go doesn't have generics so we need to do this cast
	datastoreSvc := uncastedDatastoreSvc.(*datastore.DatastoreService)

	apiSvc, err := test.addApiService(networkCtx, api1ServiceId, defaultPartitionId, datastoreSvc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v'", api1ServiceId)
	}

	if err := apiSvc.AddPerson(testPersonId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the test person in preparation for the test")
	}
	if err := apiSvc.IncrementBooksRead(testPersonId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred test person's books read in preparation for the test")
	}

	return networkCtx, nil
}


func (test NetworkPartitionTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	// Go doesn't have generics so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)


	logrus.Info("Partitioning API and datastore services off from each other...")
	blockedConnRepartitioner, err := getTwoPartitionsRepartitioner(castedNetwork, true, false)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the 2-partition repartitioner with blocked connection"))
	}
	if err := castedNetwork.RepartitionNetwork(blockedConnRepartitioner); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network to block access between API <-> datastore"))
	}
	logrus.Info("Repartition complete")

	logrus.Info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")
	uncastedApi1Service, err := castedNetwork.GetService(api1ServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the API 1 service interface"))
	}
	api1Service := uncastedApi1Service.(*api.ApiService) // Necessary because Go doesn't have generics
	if err := api1Service.IncrementBooksRead(testPersonId); err == nil {
		testCtx.Fatal(stacktrace.NewError("Expected the book increment call via API 1 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown"))
	} else {
		logrus.Infof("Incrementing books read via API 1 threw the following error as expected due to network partition: %v", err)
	}

	// Adding another API service while the partition is in place ensures that partitiong works even when you add a node
	logrus.Info("Adding second API container, to ensure adding a network under partition works...")
	uncastedDatastoreSvc, err := castedNetwork.GetService(datastoreServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the datastore service interface"))
	}
	api2Service, err := test.addApiService(
		castedNetwork,
		api2ServiceId,
		apiPartitionId,
		uncastedDatastoreSvc.(*datastore.DatastoreService))
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred adding the second API service to the network"))
	}
	logrus.Info("Second API container added successfully")

	logrus.Info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")
	if err := api2Service.IncrementBooksRead(testPersonId); err == nil {
		testCtx.Fatal(stacktrace.NewError("Expected the book increment call via API 2 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown"))
	} else {
		logrus.Infof("Incrementing books read via API 2 threw the following error as expected due to network partition: %v", err)
	}

	// Now, open the network back up
	logrus.Info("Repartitioning to heal partition between API and datastore...")
	openConnRepartitioner, err := getTwoPartitionsRepartitioner(castedNetwork, false, true)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the 2-partition repartitioner with open connection"))
	}
	if err := castedNetwork.RepartitionNetwork(openConnRepartitioner); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred healing the partition"))
	}
	logrus.Info("Partition healed successfully")

	logrus.Info("Making another call via API 1 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	if err := api1Service.IncrementBooksRead(testPersonId); err != nil {
		testCtx.Fatal(stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 1, even though the partition should have been " +
				"healed by the goroutine",
		))
	}
	logrus.Info("Successfully incremented books read via API 1, indicating that the partition has healed successfully!")

	logrus.Info("Making another call via API 2 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	if err := api2Service.IncrementBooksRead(testPersonId); err != nil {
		testCtx.Fatal(stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 2, even though the partition should have been " +
				"healed by the goroutine",
		))
	}
	logrus.Info("Successfully incremented books read via API 2, indicating that the partition has healed successfully!")
}


func (test *NetworkPartitionTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{
		IsPartitioningEnabled: true,
	}
}

func (test NetworkPartitionTest) GetExecutionTimeout() time.Duration {
	return 120 * time.Second
}

func (test NetworkPartitionTest) GetSetupTeardownBuffer() time.Duration {
	return 60 * time.Second
}

// ========================================================================================================
//                                     Private helper functions
// ========================================================================================================
func (test NetworkPartitionTest) addApiService(
		networkCtx *networks.NetworkContext,
		serviceId services.ServiceID,
		partitionId networks.PartitionID,
		datastoreSvc *datastore.DatastoreService) (*api.ApiService, error) {
	apiInitializer := api.NewApiContainerInitializer(test.apiImage, datastoreSvc)
	uncastedApiSvc, apiChecker, err := networkCtx.AddServiceToPartition(serviceId, partitionId, apiInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}
	if err := apiChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API service to start")
	}
	return uncastedApiSvc.(*api.ApiService), nil
}

/*
Creates a repartitioner that will partition the network between the API & datastore services, with the connection between them configurable
 */
func getTwoPartitionsRepartitioner(
		networkCtx *networks.NetworkContext,
		isConnectionBlocked bool,
		isApi2ServiceAddedYet bool) (*networks.Repartitioner, error){
	apiPartitionServiceIds := []services.ServiceID{
		api1ServiceId,
	}
	if isApi2ServiceAddedYet {
		apiPartitionServiceIds = append(apiPartitionServiceIds, api2ServiceId)
	}

	repartitioner, err := networkCtx.GetRepartitionerBuilder(
			false,
		).WithPartition(
			apiPartitionId,
			apiPartitionServiceIds...
		).WithPartition(
			datastorePartitionId,
			datastoreServiceId,
		).WithPartitionConnection(
			apiPartitionId,
			datastorePartitionId,
			isConnectionBlocked,
		).Build()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a three-partition repartitioner with isConnectionBlocked = %v",
			isConnectionBlocked)
	}
	return repartitioner, nil
}


