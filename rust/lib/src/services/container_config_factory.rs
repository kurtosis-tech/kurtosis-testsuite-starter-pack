use std::{collections::HashMap, path::PathBuf};
use anyhow::Result;

use super::{container_creation_config::ContainerCreationConfig, container_run_config::ContainerRunConfig, service::Service};

pub trait ContainerConfigFactory<S: Service> {
    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_creation_config(&self, container_ip_addr: &str) -> Result<ContainerCreationConfig<S>>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_run_config(&self, container_ip_addr: &str, generated_file_filepaths: HashMap<String, PathBuf>) -> Result<ContainerRunConfig>;
}