use std::time::Duration;
use anyhow::Result;
use async_trait::async_trait;

use crate::networks::{network::Network, network_context::NetworkContext};

use super::{test_configuration::TestConfiguration, test_context::TestContext};

/*
An interface encapsulating a test to run against a test network.
 */
#[async_trait]
pub trait Test {
	type N: Network + Send;

	/// Defines the configuration object that controls how the test will be executed. If you want to enable advanced
	/// features like network partitioning, you can do so here.
	fn get_test_configuration(&self) -> TestConfiguration;

	// Initializes the network to the desired state before test execution
	async fn setup(&mut self, network_ctx: NetworkContext) -> Result<Box<Self::N>>;

	/// Runs test logic against the given network, with failures reported using the given context.
	///
	/// Args:
	/// 	network: A user-defined representation of the network.
	/// 	test_ctx: The test context, which is the user's tool for making test assertions.
	async fn run(&self, network: Box<Self::N>, test_ctx: TestContext) -> Result<()>;

	/// How long the test will be given to do the pre-execution setup before the test will be
	/// 	hard-killed.
	fn get_setup_timeout(&self) -> Duration;

	/// The amount of time the test's `Run` method will be allowed to execute for before it's killed and the test
	/// is marked as failed. This does NOT include the time needed to do pre-test setup, which is handled by
	/// GetSetupTimeout.
	fn get_execution_timeout(&self) -> Duration;
}
