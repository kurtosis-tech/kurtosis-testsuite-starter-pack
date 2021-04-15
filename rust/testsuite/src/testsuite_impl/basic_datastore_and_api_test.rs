use anyhow::{anyhow, Context, Result};
use std::{time::Duration};
use kurtosis_rust_lib::{networks::network_context::NetworkContext, testsuite::{test::Test}};
use crate::services_impl::{api::{api_container_config_factory::ApiContainerConfigFactory, api_service::ApiService}, datastore::datastore_container_config_factory::DatastoreContainerConfigFactory};

const DATASTORE_SERVICE_ID_STR: &str = "datastore";
const API_SERVICE_ID: &str = "api";

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS: Duration = Duration::from_secs(1);
const WAIT_FOR_STARTUP_MAX_NUM_POLLS: u32 = 15;

const TEST_PERSON_ID: u32 = 23;
const TEST_NUM_BOOKS_READ: u64 = 3;

pub struct BasicDatastoreAndApiTest {
    datastore_image: String,
    api_image: String,
}

impl BasicDatastoreAndApiTest {
    pub fn new(datastore_image: String, api_image: String) -> BasicDatastoreAndApiTest {
        return BasicDatastoreAndApiTest{
            datastore_image,
            api_image,
        }
    }
}

impl Test for BasicDatastoreAndApiTest {
    type N = NetworkContext;

    fn configure(&self, builder: &mut kurtosis_rust_lib::testsuite::test_configuration_builder::TestConfigurationBuilder) {
        builder.with_setup_timeout_seconds(60)
            .with_run_timeout_seconds(60);
    }

    fn setup(&mut self, mut network_ctx: NetworkContext) -> Result<Box<NetworkContext>> {
        let datastore_initializer = DatastoreContainerConfigFactory::new(self.datastore_image.clone());
        let (datastore_service, datastore_checker) = network_ctx.add_service(&DATASTORE_SERVICE_ID_STR.to_owned(), &datastore_initializer)
            .context("An error occurred adding the datastore service")?;
        datastore_checker.wait_for_startup(&WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, WAIT_FOR_STARTUP_MAX_NUM_POLLS)
            .context("An error occurred waiting for the datastore service to start")?;

        let api_initializer = ApiContainerConfigFactory::new(self.api_image.clone(), &datastore_service);
        let (_, api_checker) = network_ctx.add_service(&API_SERVICE_ID.to_owned(), &api_initializer)
            .context("An error occurred adding the API service")?;
        api_checker.wait_for_startup(&WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, WAIT_FOR_STARTUP_MAX_NUM_POLLS)
            .context("An error occurred waiting for the API service to start")?;
        return Ok(Box::new(network_ctx));
    }

    fn run(&self, network: Box<NetworkContext>) -> Result<()> {
        let api_service = network.get_service::<ApiService>(&API_SERVICE_ID.to_owned())
            .context("An error occurred getting the API service")?;

        info!("Verifying that person with test ID '{}' doesn't already exist...", TEST_PERSON_ID);
        let person_or_err =  api_service.get_person(TEST_PERSON_ID);
        if person_or_err.is_ok() {
            return Err(anyhow!(
                "Expected an error trying to get a person who doesn't exist yet, but didn't receive one"
            ));
        }
        info!("Verified that test person doesn't already exist");

        info!("Adding test person with ID '{}'...", TEST_PERSON_ID);
        api_service.add_person(TEST_PERSON_ID)
            .context(format!("An error occurred adding person with test ID '{}'", TEST_PERSON_ID))?;
        info!("Test person added");

        info!("Incrementing test person's number of books read by {}...", TEST_NUM_BOOKS_READ);
        for _ in 0..TEST_NUM_BOOKS_READ {
            api_service.increment_books_read(TEST_PERSON_ID)
                .context("An error occurred incrementing the number of books read")?;
        }
        info!("Incremented number of books read");

        info!("Retrieving test person to verify number of books read...");
        let person = api_service.get_person(TEST_PERSON_ID)
            .context("An error occurred getting the test person to verify the number of books read")?;
        info!("Retrieved test person");

        if person.books_read != TEST_NUM_BOOKS_READ {
            return Err(anyhow!(
                "Expected number of book read '{}' != actual number of books read '{}'",
                TEST_NUM_BOOKS_READ,
                person.books_read,
            ));
        }
        return Ok(());
    }
}