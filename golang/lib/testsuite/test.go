/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"time"
)

/*
An interface encapsulating a test to run against a test network.
 */
type Test interface {
	/*
		Defines the configuration object that controls how the test will be executed. If you want to enable advanced
			features like network partitioning, you can do so here.
	*/
	GetTestConfiguration() TestConfiguration

	// Initializes the network to the desired state before test execution
	Setup(networkCtx *networks.NetworkContext) (networks.Network, error)

	// NOTE: if Go had generics, 'network' would be a parameterized type representing the network that this test consumes
	// as produced by the NetworkLoader
	/*
	Runs test logic against the given network, with failures reported using the given context.

	Args:
		network: A user-defined representation of the network. NOTE: Because Go doesn't have generics, this will need to
			be casted to the appropriate type.
		testCtx: The test context, which is the user's tool for making test assertions.
	 */
	Run(network networks.Network, testCtx TestContext)

	/*
	The amount of time the test's `Run` method will be allowed to execute for before it's killed and the test
		is marked as failed. This does NOT include the time needed to do pre-test setup or post-test teardown,
		which is handled by `GetSetupTeardownBuffer`. The total amount of time a test (with setup & teardown) is allowed
		to run for = GetExecutionTimeout + GetSetupBuffer.
	 */
	GetExecutionTimeout() time.Duration

	/*
	How long the test will be given to do the pre-execution setup and post-setup teardown before the test will be
		hard-killed. The total amount of time a test (with setup & teardown) is allowed to run
		for = GetExecutionTimeout + GetSetupTeardownBuffer.
	 */
	GetSetupTeardownBuffer() time.Duration
}
