use kurtosis_rust_lib::execution::test_suite_configurator::TestSuiteConfigurator;
use anyhow::Result;

pub struct ExampleTestsuiteConfigurator {}

impl ExampleTestsuiteConfigurator {
    pub fn new() -> ExampleTestsuiteConfigurator {
        return ExampleTestsuiteConfigurator{};
    }
}

impl TestSuiteConfigurator for ExampleTestsuiteConfigurator {
    fn set_log_level(&self, log_level_str: &str) -> Result<()> {
        // Nothing to do here since the log level is configured via the RUST_LOG environment variable
        return Ok(());
    }

    fn parse_params_and_create_suite(&self, params_json_str: &str) -> Result<Box<dyn kurtosis_rust_lib::testsuite::testsuite::TestSuite>> {
        todo!()
    }
}