use anyhow::Result;

use crate::testsuite::testsuite::TestSuite;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub trait TestSuiteConfigurator {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn set_log_level(&self, log_level_str: &str) -> Result<()>;

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn parse_params_and_create_suite(&self, params_json_str: &str) -> Result<Box<dyn TestSuite>>;
}
