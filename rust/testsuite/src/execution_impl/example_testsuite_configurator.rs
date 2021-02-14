use std::str::FromStr;

use kurtosis_rust_lib::{execution::test_suite_configurator::TestSuiteConfigurator};
use anyhow::{Context, Result};
use log::{LevelFilter};
use simplelog::{Config, TermLogger};

use crate::testsuite_impl::example_testsuite::ExampleTestsuite;

pub struct ExampleTestsuiteConfigurator {}

impl ExampleTestsuiteConfigurator {
    pub fn new() -> ExampleTestsuiteConfigurator {
        return ExampleTestsuiteConfigurator{};
    }
}

impl TestSuiteConfigurator for ExampleTestsuiteConfigurator {
    fn set_log_level(&self, log_level: &str) -> Result<()> {
        let level_filter = LevelFilter::from_str(log_level)
            .context(format!("Could not parse log level str '{}' to a log level filter", log_level))?;
        TermLogger::init(level_filter, Config::default(), simplelog::TerminalMode::Mixed)
            .context("An error occurred initializing the logger")?;
        return Ok(());
    }

    fn parse_params_and_create_suite(&self, params_json_str: &str) -> Result<Box<dyn kurtosis_rust_lib::testsuite::testsuite::TestSuite>> {
        // TODO actually parse params_json_str
        let suite = ExampleTestsuite::new(
            String::from("kurtosistech/example-microservices_api"),
            String::from("kurtosistech/example-microservices_datastore"),
        );
        return Ok(Box::new(suite));
    }
}