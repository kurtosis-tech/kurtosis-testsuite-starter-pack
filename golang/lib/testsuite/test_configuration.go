/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite

import "github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"

/*
Holds configuration values that, if set, give the test the ability to do special things
 */
type TestConfiguration struct {
	// If true, enables the test to set up network partitions between services
	// This should NOT be done thoughtlessly, however - when partitioning is enabled,
	//  adding services will be slower because all the other nodes in the network will
	//  need to update their iptables for the new node. The slowdown will scale with the
	//  number of services in your network.
	IsPartitioningEnabled bool

	// A mapping of ID -> URL where the artifact containing files should be downloaded from
	// The ID is the ID that service initializers will use when requesting to use the artifact
	FilesArtifactUrls map[services.FilesArtifactID]string

}
