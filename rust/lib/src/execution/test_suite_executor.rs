use std::{borrow::Borrow, collections::{HashMap}, thread::sleep, time::Duration};
use anyhow::{Context, Result, anyhow};
use log::debug;
use tokio::runtime::Runtime;
use tonic::{Request, transport::{Channel}};

use crate::{core_api_bindings::api_container_api::{SuiteAction, TestMetadata, TestSuiteMetadata, api_container_service_client::ApiContainerServiceClient, suite_metadata_serialization_service_client::SuiteMetadataSerializationServiceClient, suite_registration_service_client::SuiteRegistrationServiceClient, test_execution_service_client::TestExecutionServiceClient}, testsuite::testsuite::TestSuite};

use super::test_suite_configurator::TestSuiteConfigurator;

const GRPC_SERVER_STOP_GRACE_PERIOD: Duration = Duration::from_secs(5);
const MAX_CONNECTION_ATTEMPTS: u32 = 20;
const TIME_BETWEEN_CONNECTION_RETRIES: Duration = Duration::from_millis(500);

pub struct TestSuiteExecutor {
    kurtosis_api_socket: Option<String>, // Can be empty if testsuite is in metadata-providing mode
    log_level: String,
    params_json: String,
	configurator: Box<dyn TestSuiteConfigurator>,
}

impl TestSuiteExecutor {
    pub fn new(kurtosis_api_socket: Option<String>, log_level: &str, params_json: &str, configurator: Box<dyn TestSuiteConfigurator>) -> TestSuiteExecutor {
        return TestSuiteExecutor{
            kurtosis_api_socket,
            log_level: log_level.to_owned(),
            params_json: params_json.to_owned(),
			configurator,
        };
    }

	pub fn run(&self) -> Result<()> {
		self.configurator.set_log_level(&self.log_level)
			.context("An error occurred setting the loglevel before running the testsuite")?;

		let suite = self.configurator.parse_params_and_create_suite(&self.params_json)
			.context("An error occurred parsing the suite params JSON and creating the testsuite")?;

		let mut api_container_service: Option<ApiContainerServiceClient<Channel>> = None;
		match self.kurtosis_api_socket {
			Some(kurtosis_api_socket_str) => {
				let url = format!("http://{}", kurtosis_api_socket_str);
				// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
				let endpoint = Channel::from_shared(url)
					.context(format!("An error occurred creating the endpoint to Kurtosis API socket '{}'", kurtosis_api_socket_str))?;

				let async_runtime = Runtime::new()
					.context("An error occurred creating the Tokio async runtime")?;

				// Sometimes the API container is still in the process of starting, so we retry the channel connection a few times
				let channel: Channel;
				let mut connection_attempts: u32 = 0;
				loop {
					if connection_attempts >= MAX_CONNECTION_ATTEMPTS {
						return Err(
							anyhow!(
								"Failed to connect to API container, even after {} retries spaced {}ms apart",
								MAX_CONNECTION_ATTEMPTS,
								TIME_BETWEEN_CONNECTION_RETRIES.as_millis(),
							)
						);
					}
					let channel_or_err = async_runtime.block_on(endpoint.connect());
					match channel_or_err {
						Ok(returned_channel) => {
							channel = returned_channel;
							break;
						}
						Err(err) => {
							debug!(
								"The following error occurred connecting to the API container; retrying in {}ms: {}", 
								TIME_BETWEEN_CONNECTION_RETRIES.as_millis(),
								err.to_string()
							);
						}
					}
					sleep(TIME_BETWEEN_CONNECTION_RETRIES);
					connection_attempts += 1;
				}

				let api_container_client_channel = channel.clone(); // This *seems* weird to clone a channel, but this is apparently how Tonic wants it
				api_container_service = Some(ApiContainerServiceClient::new(api_container_client_channel));
			},
			None => {},
		}

		// TODO

		return Ok(());
	}
}