use crate::{core_api_bindings::api_container_api::{TestMetadata}};

use anyhow::{Result};
use tokio::runtime::Runtime;
use tonic::transport::Channel;
// This trait is necessary to genericize across tests, so that they can be put inside a HashMap
// since Rust doesn't allow things like HashMap<String, Test<? extends Network>>
// See: https://discord.com/channels/442252698964721669/448238009733742612/809977090740977674
pub trait DynTest {
	fn get_test_metadata(&self) -> Result<TestMetadata>;

    fn setup_and_run(&mut self, async_runtime: Runtime, channel: Channel) -> Result<()>;
}