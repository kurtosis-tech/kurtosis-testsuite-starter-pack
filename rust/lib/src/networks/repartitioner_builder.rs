use std::collections::HashSet;

use super::repartitioner_builder_actions::RepartitionerMutator;

// This struct is designed not to throw an error on any of its methods, so that they can be fluently chained together
// An error will only be thrown on "build"
pub struct RepartitionerBuilder {
	// Whether the default (unspecified) connection between partitions is blocked or not
    is_default_partition_connection_blocked: bool,
    mutators: Vec<Box<dyn RepartitionerMutator>>,
}

impl RepartitionerBuilder {
    pub (super) fn new(is_default_partition_connection_blocked: bool) -> RepartitionerBuilder {
        return RepartitionerBuilder{
            is_default_partition_connection_blocked,
            mutators: Vec::new(),
        }
    }

    pub fn with_partition(partition_id: &str, service_ids: HashSet<String>) {

    }
}

/*
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
 */