use anyhow::Result;

use crate::networks::{network::Network, network_context::NetworkContext};

use super::{test_configuration_builder::TestConfigurationBuilder};

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub trait Test {
	type N: Network;

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn configure(&self, builder: &mut TestConfigurationBuilder);

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn setup(&mut self, network_ctx: NetworkContext) -> Result<Box<Self::N>>;

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn run(&self, network: Box<Self::N>) -> Result<()>;
}
