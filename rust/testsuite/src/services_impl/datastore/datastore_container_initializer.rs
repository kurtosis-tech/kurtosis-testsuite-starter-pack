use kurtosis_rust_lib::services::{docker_container_initializer, service_context::{ServiceContext}};
use std::{collections::{HashSet, HashMap}, path::PathBuf};
use crate::services_impl::datastore::datastore_service::DatastoreService;
use std::fs::File;
use anyhow::Result;

const PORT: u32 = 1323;
const PROTOCOL: &str = "tcp";
const TEST_VOLUME_MOUNTPOINT: &str = "/test-volume";

pub struct DatastoreContainerInitializer {
    docker_image: String,
}

impl DatastoreContainerInitializer {
    pub fn new(docker_image: &str) -> DatastoreContainerInitializer {
        return DatastoreContainerInitializer{
            docker_image: docker_image.to_owned(),
        };
    }
}

impl docker_container_initializer::DockerContainerInitializer<DatastoreService> for DatastoreContainerInitializer {
    fn get_docker_image(&self) -> &str {
        return &self.docker_image;
    }

    fn get_used_ports(&self) -> HashSet<String> {
        let mut result = HashSet::new();
        result.insert(format!("{}/{}", PORT, PROTOCOL));
        return result;
    }

    fn get_service(&self, service_context: ServiceContext) -> Box<dyn kurtosis_rust_lib::services::service::Service> {
        let service = DatastoreService::new(
            service_context,
            PORT);
        return Box::new(service);
    }

    fn get_files_to_generate(&self) -> HashSet<String> {
        return HashSet::new();
    }

    fn initialize_generated_files(&self, _: HashMap<String, File>) -> Result<()> {
        return Ok(());
    }

    fn get_files_artifact_mountpoints(&self) -> HashMap<String, String> {
        return HashMap::new();
    }


    fn get_test_volume_mountpoint(&self) -> &'static str {
        return TEST_VOLUME_MOUNTPOINT;
    }

    fn get_start_command_overrides(
            &self,
            _: HashMap<String, PathBuf>, 
            _: &str
    ) -> Result<(Option<Vec<String>>, Option<Vec<String>>)> {
        // We have a launch CMD specified in the Dockerfile the datastore service was built with and we don't need
        // to specify an ENTRYPOINT, so we leave everything nil
        return Ok((None, None))
    }

}
