use std::{borrow::Borrow, rc::Rc, time::Duration};
use anyhow::{Context, Result, anyhow};

use kurtosis_rust_lib::{networks::{network::Network, network_context::NetworkContext}, services::service::ServiceId};

use crate::services_impl::{api::{api_service::ApiService, api_container_config_factory::ApiContainerConfigFactory}, datastore::{datastore_container_config_factory::DatastoreContainerConfigFactory, datastore_service::DatastoreService}};

const DATASTORE_SERVICE_ID_STR: &str = "datastore";
const API_SERVICE_ID_PREFIX: &str = "api-";

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS: Duration = Duration::from_secs(1);
const WAIT_FOR_STARTUP_MAX_NUM_POLLS: u32 = 15;

// A custom Network implementation is intended to make test-writing easier by wrapping low-level
//   NetworkContext calls with custom higher-level business logic
pub struct TestNetwork {
	network_ctx: NetworkContext,
	datastore_service_image: String,
	api_service_image: String,
    datastore_service: Option<Rc<DatastoreService>>,
    person_modifying_api_service: Option<Rc<ApiService>>,
    person_retrieving_api_service: Option<Rc<ApiService>>,
    next_api_service_id: u32,
}

impl TestNetwork {
    pub fn new(network_ctx: NetworkContext, datastore_service_image: String, api_service_image: String) -> TestNetwork {
        return TestNetwork {
            network_ctx,
            datastore_service_image,
            api_service_image,
            datastore_service: None,
            person_modifying_api_service: None,
            person_retrieving_api_service: None,
            next_api_service_id: 0,
        };
    }

    // Custom network implementations usually have a "setup" method (possibly parameterized) that is used
    //  in the Test.Setup function of each test
    pub fn setup_datastore_and_two_api_services(&mut self) -> Result<()> {
        if self.datastore_service.is_some() {
            return Err(anyhow!(
                "Cannot add datastore service to network; datastore already exists!"
            ));
        }
        if self.person_modifying_api_service.is_some() || self.person_retrieving_api_service.is_some() {
            return Err(anyhow!(
                "Cannot add API services to network; one or more API services already exists"
            ));
        }

        let config_factory = DatastoreContainerConfigFactory::new(self.datastore_service_image.clone());
        let (service, host_port_bindings, checker) = self.network_ctx.add_service(&DATASTORE_SERVICE_ID_STR.to_owned(), &config_factory)
            .context("An error occurred adding the datastore service")?;
        checker.wait_for_startup(&WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, WAIT_FOR_STARTUP_MAX_NUM_POLLS)
            .context("An error occurred waiting for the datastore service to start")?;
        info!("Added datastore service with host port bindings: {:?}", host_port_bindings);
        self.datastore_service = Some(service);

        let person_modifying_api_service = self.add_api_service()
            .context("An error occurred adding the person-modifying API service")?;
        self.person_modifying_api_service = Some(person_modifying_api_service);

        let person_retrieving_api_service = self.add_api_service()
            .context("An error occurred adding the person-retrieving API service")?;
        self.person_retrieving_api_service = Some(person_retrieving_api_service);

        return Ok(());
    }

    // Custom network implementations will also usually have getters, to retrieve information about the
    //  services created during setup
    pub fn get_person_modifying_api_service(&self) -> Result<&ApiService> {
        match &self.person_modifying_api_service {
            Some(service_box) => return Ok(service_box.borrow()),
            None => return Err(anyhow!(
                "No person-modifying API service exists"
            )),
        }
    }
    pub fn get_person_retrieving_api_service(&self) -> Result<&ApiService> {
        match &self.person_retrieving_api_service {
            Some(service_box) => return Ok(service_box.borrow()),
            None => return Err(anyhow!(
                "No person-retrieving API service exists"
            )),
        }
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    fn add_api_service(&mut self) -> Result<Rc<ApiService>> {
        let datastore;
        match &self.datastore_service {
            Some(service_box) => datastore = service_box,
            None => return Err(anyhow!(
                "Cannot add API service to network; no datastore service exists"
            )),
        }

        let service_id: ServiceId = format!("{}{}", API_SERVICE_ID_PREFIX, self.next_api_service_id);
        self.next_api_service_id += 1;

        let config_factory = ApiContainerConfigFactory::new(self.api_service_image.clone(), datastore.borrow());
        let (api_service, host_port_bindings, checker) = self.network_ctx.add_service(&service_id, &config_factory)
            .context("An error occurred adding the API service")?;
        checker.wait_for_startup(&WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, WAIT_FOR_STARTUP_MAX_NUM_POLLS)
            .context("An error occurred waiting for the API service to start")?;
        info!("Added API service with host port bindings: {:?}", host_port_bindings);
        return Ok(api_service)
    }
}

impl Network for TestNetwork {}