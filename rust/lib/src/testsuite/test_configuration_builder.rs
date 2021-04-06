use std::collections::HashMap;

use super::test_configuration::TestConfiguration;

// vvvvvvvvv Update the docs if you change these vvvvvvvvvvv
const DEFAULT_SETUP_TIMEOUT_SECONDS: u32 = 60;
const DEFAULT_RUN_TIMEOUT_SECONDS: u32 = 60;
const DEFAULT_IS_PARTITIONING_ENABLED: bool = false;
// ^^^^^^^^^ Update the docs if you change these ^^^^^^^^^^^

pub struct TestConfigurationBuilder {
    setup_timeout_seconds: u32,
    run_timeout_seconds: u32,
    is_partitioning_enabled: bool,
    files_artifact_urls: HashMap<String, String>,
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
impl TestConfigurationBuilder {
    pub fn new_test_configuration_builder() -> TestConfigurationBuilder {
        return TestConfigurationBuilder{
            setup_timeout_seconds: DEFAULT_SETUP_TIMEOUT_SECONDS,
            run_timeout_seconds: DEFAULT_RUN_TIMEOUT_SECONDS,
            is_partitioning_enabled: DEFAULT_IS_PARTITIONING_ENABLED,
            files_artifact_urls: HashMap::new(),
        }
    }

    pub fn with_setup_timeout_seconds(&mut self, setup_timeout_seconds: u32) -> &mut TestConfigurationBuilder {
        self.setup_timeout_seconds = setup_timeout_seconds;
        return self
    }

    pub fn with_run_timeout_seconds(&mut self, run_timeout_seconds: u32) -> &mut TestConfigurationBuilder {
        self.run_timeout_seconds = run_timeout_seconds;
        return self;
    }

    pub fn with_partitioning_enabled(&mut self, is_partitioning_enabled: bool) -> &mut TestConfigurationBuilder {
        self.is_partitioning_enabled = is_partitioning_enabled;
        return self
    }

    pub fn with_files_artifact_urls(&mut self, files_artifact_urls: HashMap<String, String>) -> &mut TestConfigurationBuilder {
        self.files_artifact_urls = files_artifact_urls;
        return self
    }

    pub fn build(&self) -> TestConfiguration {
        return TestConfiguration{
            test_setup_timeout_seconds: self.setup_timeout_seconds,
            test_run_timeout_seconds: self.run_timeout_seconds,
            is_partitioning_enabled: self.is_partitioning_enabled,
            files_artifact_urls: self.files_artifact_urls.clone(),
        }
    }
}