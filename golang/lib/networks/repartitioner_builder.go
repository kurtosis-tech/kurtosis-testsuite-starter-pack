/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
)


// This struct is designed not to throw an error on any of its methods, so that they can be fluently chained together
// An error will only be thrown on "build"
type RepartitionerBuilder struct {
	// Whether the default (unspecified) connection between partitions is blocked or not
	isDefaultPartitionConnectionBlocked bool
	mutators                            []repartitionerMutator
}

func newRepartitionerBuilder(isDefaultPartitionConnectionBlocked bool) *RepartitionerBuilder {
	return &RepartitionerBuilder{
		isDefaultPartitionConnectionBlocked: isDefaultPartitionConnectionBlocked,
		mutators:                            []repartitionerMutator{},
	}
}

func (builder *RepartitionerBuilder) WithPartition(partition PartitionID, services ...services.ServiceID) *RepartitionerBuilder {
	action := addPartitionAction{
		partition: partition,
		services:  services,
	}
	builder.mutators = append(builder.mutators, action)
	return builder
}

func (builder *RepartitionerBuilder) WithPartitionConnection(partitionA PartitionID, partitionB PartitionID, isBlocked bool) *RepartitionerBuilder {
	action := addPartitionConnectionAction{
		partitionA: partitionA,
		partitionB: partitionB,
		connection: &core_api_bindings.PartitionConnectionInfo{
			IsBlocked: isBlocked,
		},
	}
	builder.mutators = append(builder.mutators, action)
	return builder
}

/*
Builds a Repartitioner by applying the transformations specified on the RepartitionerBuilder
 */
func (builder *RepartitionerBuilder) Build() (*Repartitioner, error) {
	repartitioner := &Repartitioner{
		partitionServices: map[PartitionID]*serviceIdSet{},
		partitionConnections: map[PartitionID]map[PartitionID]*core_api_bindings.PartitionConnectionInfo{},
		defaultConnection: &core_api_bindings.PartitionConnectionInfo{
			IsBlocked: builder.isDefaultPartitionConnectionBlocked,
		},
	}

	for idx, mutator := range builder.mutators {
		if err := mutator.mutate(repartitioner); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred applying repartitioner builder operation #%v", idx)
		}
	}
	return repartitioner, nil
}
