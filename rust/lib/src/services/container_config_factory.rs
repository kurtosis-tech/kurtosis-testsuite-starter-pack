pub trait ContainerConfigFactory {
    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_creation_config(&self, container_ip_addr: &str) -> Result<ContainerCreationConfig>;

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    fn get_run_config(&self, container_ip_addr: &str, generated_file_filepaths: HashMap<String, String>) -> Result<ContainerRunConfig>;
}