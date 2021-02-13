use std::collections::{HashMap, HashSet};

use crate::core_api_bindings::api_container_api::PartitionConnectionInfo;

pub struct Repartitioner {
    // TODO convert key to be type alias PartitionID and value to ServiceID
    pub (super) partition_services: HashMap<String, HashSet<String>>,
    // Convert keys of both maps to be PartitionID
    pub (super) partition_connections: HashMap<String, HashMap<String, PartitionConnectionInfo>>,
    pub (super) default_connection: PartitionConnectionInfo,
}
