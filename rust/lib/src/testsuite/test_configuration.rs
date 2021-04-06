use std::collections::HashMap;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct TestConfiguration {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub test_setup_timeout_seconds: u32,

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub test_run_timeout_seconds: u32,

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub is_partitioning_enabled: bool,

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub files_artifact_urls: HashMap<String, String>,
}