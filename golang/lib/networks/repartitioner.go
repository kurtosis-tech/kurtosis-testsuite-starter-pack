/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networks

import "github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"

type PartitionID string

type Repartitioner struct {
	partitionServices map[PartitionID]*serviceIdSet
	partitionConnections map[PartitionID]map[PartitionID]*core_api_bindings.PartitionConnectionInfo
	defaultConnection *core_api_bindings.PartitionConnectionInfo
}
