/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
)

type repartitionerMutator interface {
	mutate(repartitioner *Repartitioner) error
}

// ======================================================================================================
//                                         Add partition
// ======================================================================================================
type addPartitionAction struct {
	partition PartitionID
	services  []services.ServiceID
}

func (a addPartitionAction) mutate(repartitioner *Repartitioner) error {
	newPartition := a.partition

	newPartitionServices := newServiceIdSet()
	for _, id := range a.services {
		newPartitionServices.add(id)
	}
	repartitioner.partitionServices[newPartition] = newPartitionServices
	return nil
}

// ======================================================================================================
//                                     Add partition connection
// ======================================================================================================
type addPartitionConnectionAction struct {
	partitionA PartitionID
	partitionB PartitionID
	connection *core_api_bindings.PartitionConnectionInfo
}

func (a addPartitionConnectionAction) mutate(repartitioner *Repartitioner) error {
	partitionA := a.partitionA
	partitionB := a.partitionB
	connectionInfo := a.connection

	partitionAConns, found := repartitioner.partitionConnections[partitionA]
	if !found {
		partitionAConns = map[PartitionID]*core_api_bindings.PartitionConnectionInfo{}
	}
	partitionAConns[partitionB] = connectionInfo
	repartitioner.partitionConnections[partitionA] = partitionAConns
	return nil
}
