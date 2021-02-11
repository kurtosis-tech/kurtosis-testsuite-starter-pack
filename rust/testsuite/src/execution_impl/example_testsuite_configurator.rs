use kurtosis_rust_lib::{execution::test_suite_configurator::TestSuiteConfigurator, testsuite::testsuite::TestSuite};
use anyhow::Result;

use crate::testsuite_impl::example_testsuite::ExampleTestsuite;

pub struct ExampleTestsuiteConfigurator {}

impl ExampleTestsuiteConfigurator {
    pub fn new() -> ExampleTestsuiteConfigurator {
        return ExampleTestsuiteConfigurator{};
    }
}

impl TestSuiteConfigurator for ExampleTestsuiteConfigurator {
    fn set_log_level(&self, _: &str) -> Result<()> {
        pretty_env_logger::init();
        return Ok(());
    }

    fn parse_params_and_create_suite(&self, params_json_str: &str) -> Result<Box<dyn kurtosis_rust_lib::testsuite::testsuite::TestSuite>> {
        // TODO use params_json_str
        return Ok(Box::new(ExampleTestsuite::new()));
    }
}