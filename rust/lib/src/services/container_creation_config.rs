use std::{collections::{HashMap, HashSet}, fs::File, io::Write};
use anyhow::Result;

use super::{service::Service, service_context::ServiceContext};

pub type ServiceCreatingFunc<S> = fn(ServiceContext) -> S;

pub type FileGeneratingFunc = fn(File) -> Result<()>;

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct ContainerCreationConfig<S: Service> {
    image: String,
    test_volume_mountpoint: String,
    used_ports_set: HashSet<String>,
    service_creating_func: ServiceCreatingFunc<S>,
    file_generating_funcs: HashMap<String, FileGeneratingFunc>,
    files_artifact_mountpoints: HashMap<String, String>
}

impl<S: Service> ContainerCreationConfig<S> {
    pub fn get_image(&self) -> &str {
        return &self.image;
    }

    pub fn get_test_volume_mountpoint(&self) -> &str {
        return &self.test_volume_mountpoint;
    }

    pub fn get_used_ports(&self) -> &HashSet<String> {
        return &self.used_ports_set;
    }

    pub fn get_service_creating_func(&self) -> &ServiceCreatingFunc<S> {
        return &self.service_creating_func;
    }

    pub fn get_file_generating_funcs(&self) -> &HashMap<String, FileGeneratingFunc> {
        return &self.file_generating_funcs;
    }

    pub fn get_files_artifact_mountpoints(&self) -> &HashMap<String, String> {
        return &self.files_artifact_mountpoints;
    }
}


// ====================================================================================================
//                                        Builder
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct ContainerCreationConfigBuilder<S: Service> {
    image: String,
    test_volume_mountpoint: String,
    used_ports: HashSet<String>,
    service_creating_func: ServiceCreatingFunc<S>,
    files_generating_funcs: HashMap<String, FileGeneratingFunc>,
    files_artifact_mountpoints: HashMap<String, String>
}

impl<S: Service> ContainerCreationConfigBuilder<S> {
    pub fn new(image: String, test_volume_mountpoint: String, service_creating_func: ServiceCreatingFunc<S>) -> ContainerCreationConfigBuilder<S> {
        return ContainerCreationConfigBuilder{
            image,
            test_volume_mountpoint,
            used_ports: HashSet::new(),
            service_creating_func,
            files_generating_funcs: HashMap::new(),
            files_artifact_mountpoints: HashMap::new(),
        }
    }

    pub fn with_used_ports(&mut self, used_ports: HashSet<String>) -> &mut ContainerCreationConfigBuilder<S> {
        self.used_ports = used_ports;
        return self;
    }

    pub fn with_generated_files(&mut self, file_generating_funcs: HashMap<String, FileGeneratingFunc>) -> &mut ContainerCreationConfigBuilder<S> {
        self.files_generating_funcs = file_generating_funcs;
        return self;
    }

    pub fn with_files_artifacts(&mut self, files_artifact_mountpoints: HashMap<String, String>) -> &mut ContainerCreationConfigBuilder<S> {
        self.files_artifact_mountpoints = files_artifact_mountpoints;
        return self;
    }

    pub fn build(&self) -> ContainerCreationConfig<S> {
        return ContainerCreationConfig{
            image: self.image.clone(),
            test_volume_mountpoint: self.test_volume_mountpoint.clone(),
            used_ports_set: self.used_ports.clone(),
            service_creating_func: self.service_creating_func.clone(),
            file_generating_funcs: self.files_generating_funcs.clone(),
            files_artifact_mountpoints: self.files_artifact_mountpoints.clone(),
        }
    }

}