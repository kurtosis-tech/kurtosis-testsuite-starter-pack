use anyhow::{anyhow, Context, Result};

use kurtosis_rust_lib::{networks::network_context::NetworkContext, services::service::ServiceId, testsuite::{test::Test}};

use crate::networks_impl::test_network::TestNetwork;

const TEST_PERSON_ID: u32 = 46;

pub struct AdvancedNetworkTest {
    datastore_service_image: String,
    api_service_image: String,

    person_modifying_api_service_id: Option<ServiceId>,
    person_retrieving_api_service_id: Option<ServiceId>,
}

impl AdvancedNetworkTest {
    pub fn new(datastore_service_image: String, api_service_image: String) -> AdvancedNetworkTest {
        return AdvancedNetworkTest{
            datastore_service_image,
            api_service_image,
            person_modifying_api_service_id: None,
            person_retrieving_api_service_id: None,
        }
    }
}

impl Test for AdvancedNetworkTest {
    type N = TestNetwork;

    fn configure(&self, builder: &mut kurtosis_rust_lib::testsuite::test_configuration_builder::TestConfigurationBuilder) {
        builder.with_setup_timeout_seconds(60)
            .with_run_timeout_seconds(60);
    }

    fn setup(&mut self, network_ctx: NetworkContext) -> Result<Box<TestNetwork>> {
        let mut network = TestNetwork::new(network_ctx, self.datastore_service_image.clone(), self.api_service_image.clone());

        network.add_datastore()
            .context("An error occurred adding the datastore")?;

        let person_modifying_api_service_id = network.add_api_service()
            .context("An error occurred adding the person-modifying API service")?;
        self.person_modifying_api_service_id = Some(person_modifying_api_service_id);

        let person_retrieving_api_service_id = network.add_api_service()
            .context("An error occurred adding the person-retrieving API service")?;
        self.person_retrieving_api_service_id = Some(person_retrieving_api_service_id);

        return Ok(Box::new(network));
    }

    fn run(&self, network: Box<TestNetwork>) -> anyhow::Result<()> {
        let person_modifying_service_id;
        match self.person_modifying_api_service_id {
            Some(ref service_id) => person_modifying_service_id = service_id,
            None => return Err(anyhow!(
                "No person-modifying service ID exists; this is a code bug"
            )),
        };
        let person_modifier = network.get_api_service(person_modifying_service_id)
            .context("An error occurred getting the person-modifying API service")?;


        let person_retrieving_service_id;
        match self.person_retrieving_api_service_id {
            Some(ref service_id) => person_retrieving_service_id = service_id,
            None => return Err(anyhow!(
                "No person-retrieving service ID exists; this is a code bug"
            )),
        }
        let person_retriever = network.get_api_service(person_retrieving_service_id)
            .context("An error occurred getting the person-retrieving API service")?;

        info!("Adding test person via person-modifying API service...");
        person_modifier.add_person(TEST_PERSON_ID)
            .context("An error occurred adding test person")?;
        info!("Test person added");

        info!("Increment test person's number of books read through person-modifying API service...");
        person_modifier.increment_books_read(TEST_PERSON_ID)
            .context("An error occurred incrementing the number of books read")?;
        info!("Incremented number of books read");

        info!("Retrieving test person to verify number of books read person-retrieving API service...");
        let person = person_retriever.get_person(TEST_PERSON_ID)
            .context("An error occurred getting the test person")?;
        info!("Retrieved test person");

        if person.books_read != 1 {
            return Err(anyhow!(
                "Expected number of books read to be incremented, but was '{}'",
                person.books_read,
            ));
        }
        return Ok(());
    }
}