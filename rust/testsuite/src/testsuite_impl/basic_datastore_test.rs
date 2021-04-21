use std::{borrow::BorrowMut, rc::Rc, time::Duration};
use anyhow::{anyhow, Context, Result};

use kurtosis_rust_lib::{networks::network_context::NetworkContext, testsuite::{test::Test}};

use crate::services_impl::datastore::{datastore_container_config_factory::DatastoreContainerConfigFactory, datastore_service::DatastoreService};

const DATASTORE_SERVICE_ID_STR: &str = "datastore";

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS: Duration = Duration::from_secs(1);
const WAIT_FOR_STARTUP_MAX_POLLS: u32 = 15;

const TEST_KEY: &str = "test-key";
const TEST_VALUE: &str = "test-value";

pub struct BasicDatastoreTest {
    datastore_image: String,
}

impl BasicDatastoreTest {
    pub fn new(datastore_image: &str) -> BasicDatastoreTest {
        return BasicDatastoreTest{
            datastore_image: datastore_image.to_owned(),
		};
    }
}

impl Test for BasicDatastoreTest {
	type N = NetworkContext;

    fn configure(&self, builder: &mut kurtosis_rust_lib::testsuite::test_configuration_builder::TestConfigurationBuilder) {
		builder.with_setup_timeout_seconds(60)
			.with_run_timeout_seconds(60);
    }

    fn setup(&mut self, mut network_ctx: NetworkContext) -> Result<Box<NetworkContext>> {
		let initializer = DatastoreContainerConfigFactory::new(self.datastore_image.clone());
		let (_, host_port_bindings, availability_checker) = network_ctx.borrow_mut().add_service(&DATASTORE_SERVICE_ID_STR.to_owned(), &initializer)
			.context("An error occurred adding the datastore service")?;
		availability_checker.wait_for_startup(&WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, WAIT_FOR_STARTUP_MAX_POLLS)
			.context("An error occurred waiting for the datastore service to become available")?;
		info!("Added datastore service with host port bindings: {:?}", host_port_bindings);
		return Ok(Box::new(network_ctx));
	}

    fn run(&self, network: Box<NetworkContext>) -> Result<()> {
		let service: Rc<DatastoreService> = network.get_service(&DATASTORE_SERVICE_ID_STR.to_owned())
			.context("An error occurred getting the datastore service")?;
		info!("Verifying that key '{}' doesn't already exist...", TEST_KEY);
		let does_exist = service.exists(TEST_KEY)
			.context(format!("An error occurred checking if key '{}' exists", TEST_KEY))?;
		if does_exist {
			return Err(anyhow!("Test key should not exist yet"));
		}
		info!("Confirmed that key '{}' doesn't already exist", TEST_KEY);

		info!("Inserting value '{}' at key '{}'...", TEST_KEY, TEST_VALUE);
		service.upsert(TEST_KEY, TEST_VALUE)
			.context(format!("An error occurred upserting value '{}' at key '{}'", TEST_VALUE, TEST_KEY))?;
		info!("Inserted value successfully");

		info!("Getting the key we just inserted to verify the value...");
		let value = service.get(TEST_KEY)
			.context(format!("An error occurred getting value for key '{}'", TEST_KEY))?;
		if value != TEST_VALUE {
			return Err(anyhow!(
				"Returned value '{}' != test value '{}'",
				value,
				TEST_VALUE,
			));
		}
		// TODO induce panic and ensure we recover from it!
		info!("Value verified");
		return Ok(());
    }
}