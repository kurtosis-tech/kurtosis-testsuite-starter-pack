

// The ID of an artifact containing files that should be mounted into a service container
// type FilesArtifactID string

use std::{collections::{HashSet, HashMap}, path::PathBuf};
use crate::services::service::Service;
use std::fs::File;
use anyhow::Result;

use super::service_context::ServiceContext;


// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub trait DockerContainerInitializer<S: Service> {
    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_docker_image(&self) -> &str;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_used_ports(&self) -> HashSet<String>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_service(&self, service_ctx: ServiceContext) -> Box<S>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_files_to_generate(&self) -> HashSet<String>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn initialize_generated_files(&self, generated_files: HashMap<String, File>) -> Result<()>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_files_artifact_mountpoints(&self) -> HashMap<String, String>;


    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_test_volume_mountpoint(&self) -> &'static str;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_start_command_overrides(
        &self,
        generated_file_filepaths: HashMap<String, PathBuf>,
        ip_addr: &str
    ) -> Result<(Option<Vec<String>>, Option<Vec<String>>)>;
}
