use std::collections::HashMap;

/*
Holds configuration values that, if set, give the test the ability to do special things
 */
pub struct TestConfiguration {
	// If true, enables the test to set up network partitions between services
	// This should NOT be done thoughtlessly, however - when partitioning is enabled,
	//  adding services will be slower because all the other nodes in the network will
	//  need to update their iptables for the new node. The slowdown will scale with the
	//  number of services in your network.
    pub is_partitioning_enabled: bool,

	// A mapping of ID -> URL where the artifact containing files should be downloaded from
	// The ID is the ID that service initializers will use when requesting to use the artifact
    // TODO CONVERT KEY TO FILESARTIFACTID TYPE
    pub files_artifact_urls: HashMap<String, String>,
}