use std::{collections::{HashMap}, thread::sleep, time::Duration};
use anyhow::{Context, Result, anyhow};
use futures::executor::block_on;
use log::debug;
use tonic::{Request, transport::{Channel}};

use crate::{core_api_bindings::api_container_api::{SuiteAction, TestMetadata, TestSuiteMetadata, suite_metadata_serialization_service_client::SuiteMetadataSerializationServiceClient, suite_registration_service_client::SuiteRegistrationServiceClient, test_execution_service_client::TestExecutionServiceClient}, testsuite::testsuite::TestSuite};

use super::test_suite_configurator::TestSuiteConfigurator;

const MAX_CONNECTION_ATTEMPTS: u32 = 20;
const TIME_BETWEEN_CONNECTION_RETRIES: Duration = Duration::from_millis(500);
const CHANNEL_TIMEOUT_SECONDS: u64 = 30;

pub struct TestSuiteExecutor {
    kurtosis_api_socket: String,
    log_level: String,
    params_json: String,
	configurator: Box<dyn TestSuiteConfigurator>,
}

impl TestSuiteExecutor {
    pub fn new(kurtosis_api_socket: &str, log_level: &str, params_json: &str, configurator: Box<dyn TestSuiteConfigurator>) -> TestSuiteExecutor {
        return TestSuiteExecutor{
            kurtosis_api_socket: kurtosis_api_socket.to_owned(),
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

		let url = format!("http://{}", self.kurtosis_api_socket);
		let endpoint = Channel::from_shared(url)
			.context(format!("An error occurred creating the endpoint to Kurtosis API socket '{}'", &self.kurtosis_api_socket))?
			// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
			.timeout(Duration::new(CHANNEL_TIMEOUT_SECONDS, 0));

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
			let channel_or_err = block_on(endpoint.connect());
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

		let suite_registration_channel = channel.clone(); // This *seems* weird to clone a channel, but this is apparently how Tonic wants it
		let mut suite_registration_client = SuiteRegistrationServiceClient::new(suite_registration_channel);
		let suite_registration_resp = block_on(suite_registration_client.register_suite(()))
			.context("An error occurred registering the testsuite container with the API container")?
			.into_inner();

		let action_int = suite_registration_resp.suite_action;
		let action = SuiteAction::from_i32(action_int)
			.context(format!("Could not convert suite action int '{}' to enum", action_int))?;
		match action {
			SuiteAction::SerializeSuiteMetadata => {
				TestSuiteExecutor::run_serialize_suite_metadata_flow(suite, channel.clone())
					.context("An error occurred running the suite metadata serialization flow")?;
			}
			SuiteAction::ExecuteTest => {
				TestSuiteExecutor::run_test_execution_flow(suite, channel.clone())
					.context("An error occurred running the test execution flow")?;
			}
		}

		return Ok(());
	}

	fn run_serialize_suite_metadata_flow(testsuite: Box<dyn TestSuite>, channel: Channel) -> Result<()> {
		let mut all_test_metadata: HashMap<String, TestMetadata> = HashMap::new();
		for (test_name, test) in testsuite.get_tests() {
			let test_metadata = test.get_test_metadata()
				.context(format!("An error occurred getting metadata for test '{}'", test_name))?;
			all_test_metadata.insert(test_name, test_metadata);
		}

		let network_width_bits = testsuite.get_network_width_bits();
		let testsuite_metadata = TestSuiteMetadata{
		    test_metadata: all_test_metadata,
		    network_width_bits: network_width_bits,
		};

		let mut client = SuiteMetadataSerializationServiceClient::new(channel);
		let req = Request::new(testsuite_metadata);
		block_on(client.serialize_suite_metadata(req))
			.context("An error occurred sending the suite metadata to the Kurtosis API server")?;
		return Ok(());
	}

	fn run_test_execution_flow(testsuite: Box<dyn TestSuite>, channel: Channel) -> Result<()> {
		let mut client = TestExecutionServiceClient::new(channel.clone());
		let resp_or_err = block_on(client.get_test_execution_info(()));
		let test_ex_info = resp_or_err
			.context("An error occurred getting the test execution info")?
			.into_inner();
		let test_name = test_ex_info.test_name;

		let mut all_tests = testsuite.get_tests();
		let test = all_tests.get_mut(&test_name)
			.context(format!(
				"Testsuite was directed to execute test '{}', but no test with that name exists in the testsuite; this is a Kurtosis code bug",
				test_name
			))?;

		// TODO TODO TODO wrap this entire thing with panic-catching
		test.setup_and_run(channel)
			.context(format!("An error occurred setting up & executing test '{}'", &test_name))?;

		return Ok(());
	}
}