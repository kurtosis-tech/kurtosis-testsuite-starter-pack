use std::collections::HashSet;

use anyhow::Result;

use super::repartitioner::Repartitioner;

pub (super) trait RepartitionerMutator {
    fn mutate(&self, repartitioner: &mut Repartitioner) -> Result<()>;
}

// ======================================================================================================
//                                         Add partition
// ======================================================================================================
struct AddPartitionAction {
    // TODO change to PartitionID custom type
    partition_id: String,
    services: HashSet<String>,
}

impl RepartitionerMutator for AddPartitionAction {
    fn mutate(&self, repartitioner: &mut Repartitioner) -> Result<()> {
        repartitioner.partition_services.insert(self.partition_id.clone(), self.services.clone());
        return Ok(());
    }
}

// ======================================================================================================
//                                     Add partition connection
// ======================================================================================================
struct AddPartitionConnectionAction {

}

/*

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
 */