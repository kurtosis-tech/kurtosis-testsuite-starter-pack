use std::{error::Error, time::Duration};

use crate::networks::{network::Network, network_context::NetworkContext};

use super::{test_configuration::TestConfiguration, test_context::TestContext};

/*
An interface encapsulating a test to run against a test network.
 */
pub trait Test<N> where N: Network + ?Sized {
	/// Defines the configuration object that controls how the test will be executed. If you want to enable advanced
	/// features like network partitioning, you can do so here.
	fn get_test_configuration(&self) -> TestConfiguration;

	// Initializes the network to the desired state before test execution
	fn setup(&self, network_ctx: NetworkContext) -> Result<Box<N>, Box<dyn Error>>;

	/// Runs test logic against the given network, with failures reported using the given context.
	///
	/// Args:
	/// 	network: A user-defined representation of the network.
	/// 	test_ctx: The test context, which is the user's tool for making test assertions.
	fn run(&self, network: N, test_ctx: TestContext);

	/// The amount of time the test's [Test::run] method will be allowed to execute for before it's killed and the test
 	/// is marked as failed. This does NOT include the time needed to do pre-test setup or post-test teardown,
 	/// which is handled by [Test::get_setup_teardown_buffer]. 
	/// 
	/// The total amount of time a test (with setup & teardown) is allowed to run for = execution 
	/// timeout + setup/teardown buffer.
	fn get_execution_timeout(&self) -> Duration;

	/// How long the test will be given to do the pre-execution setup and post-setup teardown before the test will be
	/// hard-killed. The total amount of time a test (with setup & teardown) is allowed to run
	/// for = GetExecutionTimeout + GetSetupTeardownBuffer.
	fn get_setup_teardown_buffer(&self) -> Duration;
}
