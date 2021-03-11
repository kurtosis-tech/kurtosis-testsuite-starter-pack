use crate::{core_api_bindings::api_container_api::{TestMetadata}};

use async_trait::async_trait;
use anyhow::{Result};
use tonic::transport::Channel;
// This trait is necessary to genericize across tests, so that they can be put inside a HashMap
// since Rust doesn't allow things like HashMap<String, Test<? extends Network>>
// See: https://discord.com/channels/442252698964721669/448238009733742612/809977090740977674
// NOTE: The async_trait macro is required due to:
// https://rust-lang.github.io/async-book/07_workarounds/05_async_in_traits.html
#[async_trait]
pub trait DynTest {
	fn get_test_metadata(&self) -> Result<TestMetadata>;

    async fn setup_and_run(&mut self, channel: Channel) -> Result<()>;
}