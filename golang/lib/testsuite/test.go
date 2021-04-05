/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"time"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type Test interface {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetTestConfiguration() TestConfiguration

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	Setup(networkCtx *networks.NetworkContext) (networks.Network, error)

	// NOTE: if Go had generics, 'network' would be a parameterized type representing the network that this test consumes
	// as produced by the NetworkLoader
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	Run(network networks.Network) error

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetSetupTimeout() time.Duration

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetExecutionTimeout() time.Duration

}
