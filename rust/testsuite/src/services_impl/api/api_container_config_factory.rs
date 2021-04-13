use std::{collections::{HashMap, HashSet}, fs::File};
use anyhow::{Context, Result};

use kurtosis_rust_lib::{core_api_bindings::api_container_api::FileGenerationOptions, services::{container_config_factory::ContainerConfigFactory, container_creation_config::{ContainerCreationConfig, ContainerCreationConfigBuilder, FileGeneratingFunc}, service_context::ServiceContext}};

use crate::services_impl::datastore::datastore_service::DatastoreService;

use super::api_service::ApiService;
use serde::{Deserialize, Serialize};

const PORT: u32 = 2434;
const CONFIG_FILE_KEY: &str = "config-file";
const TEST_VOLUME_MOUNTPOINT: &str = "/test-volume";

#[derive(Serialize, Deserialize, Debug)]
struct Config {
    #[serde(rename = "datastoreIp")]
    datastore_ip: String,

    #[serde(rename = "datastorePort")]
    datastore_port: u32,
}

struct ApiContainerConfigFactory<'obj> {
    image: String,
    datastore: &'obj DatastoreService,
}

impl<'obj> ApiContainerConfigFactory<'obj> {
    pub fn new(image: String, datastore: &'obj DatastoreService) -> ApiContainerConfigFactory {
        return ApiContainerConfigFactory{
            image,
            datastore,
        }
    }

    fn create_service(service_ctx: ServiceContext) -> ApiService {
        return ApiService::new(service_ctx, PORT);
    }

    fn initialize_config_file(&self, fp: File) -> Result<()> {
        debug!("Datastore IP: {} , port: {}", self.datastore.get_ip_address(), self.datastore.get_port());
        let config_obj = Config{
            datastore_ip: self.datastore.get_ip_address().to_owned(),
            datastore_port: self.datastore.get_port(),
        };
        debug!("Config obj: {:?}", config_obj);

        serde_json::to_writer(fp, &config_obj)
            .context("An error occurred serializing the config to JSON")?;

        return Ok(());
    }
}

impl<'obj> ContainerConfigFactory<ApiService> for ApiContainerConfigFactory<'obj> {
    fn get_creation_config(&self, container_ip_addr: &str) -> anyhow::Result<ContainerCreationConfig<ApiService>> {
        let mut ports = HashSet::new();
        ports.insert(format!("{}/tcp", PORT));

        let config_initialization_func = |fp: File| -> Result<()> {
            debug!("Datastore IP: {} , port: {}", self.datastore.get_ip_address(), self.datastore.get_port());
            let config_obj = Config{
                datastore_ip: self.datastore.get_ip_address().to_owned(),
                datastore_port: self.datastore.get_port(),
            };
            debug!("Config obj: {:?}", config_obj);

            serde_json::to_writer(fp, &config_obj)
                .context("An error occurred serializing the config to JSON")?;

            return Ok(());
        };

        let mut file_generation_funcs: HashMap<String, FileGeneratingFunc> = HashMap::new();
        file_generation_funcs.insert(CONFIG_FILE_KEY.to_owned(), config_initialization_func);



        ContainerCreationConfigBuilder::new(self.image, TEST_VOLUME_MOUNTPOINT, ApiContainerConfigFactory::create_service)
            .with_used_ports(ports)
            .with_generated_files(file_generating_funcs)


    }

    fn get_run_config(&self, container_ip_addr: &str, generated_file_filepaths: std::collections::HashMap<String, std::path::PathBuf>) -> anyhow::Result<kurtosis_rust_lib::services::container_run_config::ContainerRunConfig> {
        todo!()
    }
}