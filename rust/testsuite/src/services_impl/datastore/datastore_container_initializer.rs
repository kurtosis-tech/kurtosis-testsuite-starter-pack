use kurtosis_rust_lib::services::docker_container_initializer;
use std::{collections::{HashSet, HashMap}, error::Error};
use crate::services_impl::datastore::datastore_service::DatastoreService;
use std::fs::File;

const PORT: u32 = 1323;
const PROTOCOL: &str = "tcp";
const TEST_VOLUME_MOUNTPOINT: &str = "/test-volume";

struct DatastoreContainerInitializer {
    docker_image: String,
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

    fn get_service(&self, service_id: &str, ip_addr: &str) -> DatastoreService {
        return DatastoreService::new(
            service_id,
            ip_addr, 
            PORT
        );
    }

    fn get_files_to_mount() -> HashSet<String> {
        return HashSet::new();
    }

    fn initialize_mounted_files(_: HashMap<String, File>) -> Result<(), Box<dyn Error>> {
        return Ok(());
    }

    fn get_files_artifact_mountpoints() -> HashMap<String, String> {
        return HashMap::new();
    }


    fn get_test_volume_mountpoint() -> &'static str {
        return TEST_VOLUME_MOUNTPOINT;
    }

    fn get_start_command(
            _: HashMap<String, String>, 
            _: &str
    ) -> Result<Option<Vec<String>>, Box<dyn Error>> {
        // TODO change return type???
        return Ok(None)
    }
}
