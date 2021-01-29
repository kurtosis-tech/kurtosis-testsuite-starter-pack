/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	partition1 PartitionID = "partition1"
	partition2 PartitionID = "partition2"

	service1 services.ServiceID = "service1"
	service2 services.ServiceID = "service2"

	isTestRepartitionerDefaultConnBlocked = false
)

func TestAddPartitionAction(t *testing.T) {
	repartitioner := getTestRepartitioner()

	newPartition := PartitionID("new-partition")
	newService1 := services.ServiceID("new-service1")
	newService2 := services.ServiceID("new-service2")
	action := addPartitionAction{
		partition: newPartition,
		services: []services.ServiceID{newService1, newService2},
	}

	if err := action.mutate(repartitioner); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred applying the action"))
	}

	updatedPartitionServices := repartitioner.partitionServices
	assert.Equal(t, 3, len(updatedPartitionServices))
	newPartitionServices, foundNewPartition := updatedPartitionServices[newPartition]
	assert.True(t, foundNewPartition)

	assert.True(t, newPartitionServices.contains(newService1))
	assert.True(t, newPartitionServices.contains(newService2))
}

func TestAddPartitionConnectionAction(t *testing.T) {
	repartitioner := getTestRepartitioner()

	action := addPartitionConnectionAction{
		partitionA: partition1,
		partitionB: partition2,
		connection: &bindings.PartitionConnectionInfo{IsBlocked: !isTestRepartitionerDefaultConnBlocked},
	}

	if err := action.mutate(repartitioner); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred applying the action"))
	}

	updatedPartitionConnections := repartitioner.partitionConnections
	partition1Conns, found := updatedPartitionConnections[partition1]
	if !found {
		t.Fatal(stacktrace.NewError("Expected to find one connection for partition '%v'", partition1))
	}
	partition1To2Conn, found := partition1Conns[partition2]
	if !found {
		t.Fatal(stacktrace.NewError("Expected to find one connection from partition '%v' to '%v'", partition1, partition2))
	}
	assert.Equal(t, !isTestRepartitionerDefaultConnBlocked, partition1To2Conn.IsBlocked)
}

func getTestRepartitioner() *Repartitioner {
	return &Repartitioner{
		partitionServices:    map[PartitionID]*serviceIdSet{
			partition1: newServiceIdSet(service1),
			partition2: newServiceIdSet(service2),
		},
		partitionConnections: map[PartitionID]map[PartitionID]*bindings.PartitionConnectionInfo{},
		defaultConnection:    &bindings.PartitionConnectionInfo{IsBlocked: isTestRepartitionerDefaultConnBlocked},
	}
}
