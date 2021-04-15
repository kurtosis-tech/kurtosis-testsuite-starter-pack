use std::{collections::HashSet, sync::Arc};

use kurtosis_rust_lib::services::{container_config_factory::ContainerConfigFactory, container_creation_config::{ContainerCreationConfigBuilder}, container_run_config::{ContainerRunConfigBuilder}, service_context::ServiceContext};

use super::datastore_service::DatastoreService;


const PORT: u32 = 1323;
const PROTOCOL: &str = "tcp";
const TEST_VOLUME_MOUNTPOINT: &str = "/test-volume";

pub struct DatastoreContainerConfigFactory {
    image: String,
}

impl DatastoreContainerConfigFactory {
    pub fn new(image: String) -> DatastoreContainerConfigFactory {
        return DatastoreContainerConfigFactory{
            image,
        }
    }

    pub fn create_service(service_ctx: ServiceContext) -> DatastoreService {
        return DatastoreService::new(service_ctx, PORT)
    }
}

impl ContainerConfigFactory<DatastoreService> for DatastoreContainerConfigFactory {
    fn get_creation_config(&self, _container_ip_addr: &str) -> anyhow::Result<kurtosis_rust_lib::services::container_creation_config::ContainerCreationConfig<DatastoreService>> {
        let mut ports = HashSet::new();
        ports.insert(format!("{}/{}", PORT, PROTOCOL));

        let result = ContainerCreationConfigBuilder::new(
                self.image.clone(), 
                TEST_VOLUME_MOUNTPOINT.to_owned(), 
                Arc::new(DatastoreContainerConfigFactory::create_service))
            .with_used_ports(ports)
            .build();

        return Ok(result);
    }

    fn get_run_config(&self, _container_ip_addr: &str, _generated_file_filepaths: std::collections::HashMap<String, std::path::PathBuf>) -> anyhow::Result<kurtosis_rust_lib::services::container_run_config::ContainerRunConfig> {
        let result = ContainerRunConfigBuilder::new().build();
        return Ok(result);
    }
}