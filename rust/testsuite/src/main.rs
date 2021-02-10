mod execution_impl;
mod services_impl;
mod testsuite_impl;

extern crate pretty_env_logger;
#[macro_use] extern crate log;

use std::process::exit;

use clap::{App, Arg, ArgMatches};
use execution_impl::example_testsuite_configurator::ExampleTestsuiteConfigurator;
use kurtosis_rust_lib::execution::test_suite_executor::TestSuiteExecutor;
use crate::services_impl::datastore::datastore_service::DatastoreService;

const CUSTOM_PARAMS_JSON_FLAG: &str = "custom-params-json";
const KURTOSIS_API_SOCKET_FLAG: &str  = "kurtosis-api-socket";
const LOG_LEVEL_FLAG: &str = "log-level";
const FAILURE_EXIT_CODE: i32 = 1;
const SUCCESS_EXIT_CODE: i32 = 0;

fn main() {
    let matches = App::new("My Super Program")
        .arg(Arg::new(CUSTOM_PARAMS_JSON_FLAG)
            .long(CUSTOM_PARAMS_JSON_FLAG)
            .about("JSON string containing custom data that the testsuite will deserialize to modify runtime behaviour")
            .takes_value(true)
            .value_name("JSON")
            .default_value("{}"))
        .arg(Arg::new(KURTOSIS_API_SOCKET_FLAG)
            .long(KURTOSIS_API_SOCKET_FLAG)
            .about("Socket in the form of address:port of the Kurtosis API container")
            .required(true)
            .takes_value(true)
            .value_name("IP:PORT"))
        .arg(Arg::new(LOG_LEVEL_FLAG)
            .long(LOG_LEVEL_FLAG)
            .about("String indicating the loglevel that the test suite should output with")
            .required(true)
            .takes_value(true)
            .value_name("LEVEL"))
        .get_matches();

    let custom_params_json = get_arg_value(&matches, CUSTOM_PARAMS_JSON_FLAG);
    let kurtosis_api_socket = get_arg_value(&matches, KURTOSIS_API_SOCKET_FLAG);
    let log_level = get_arg_value(&matches, LOG_LEVEL_FLAG);

    // >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<
	let configurator = ExampleTestsuiteConfigurator::new();
	// >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<
    
    let configurator_box = Box::from(configurator);
    let executor = TestSuiteExecutor::new(
        kurtosis_api_socket,
        log_level,
        custom_params_json,
        configurator_box
    );
    let run_result = executor.run();
    if run_result.is_err() {
        let err = run_result.unwrap_err();
        // The {:#} incantation is how you display the full error info of an Anyhow error, as per
        // https://docs.rs/anyhow/1.0.26/anyhow/struct.Error.html
        error!("An error occurred running the test suite executor: {:#}", err);
        exit(FAILURE_EXIT_CODE);
    }
    exit(SUCCESS_EXIT_CODE);
}

fn get_arg_value<'a>(matches: &'a ArgMatches, arg_name: &'static str) -> &'a str {
    let arg_opt = matches.value_of(arg_name);
    if arg_opt.is_none() {
        error!("No argument '{}' supplied, even though it's required", arg_name);
        exit(FAILURE_EXIT_CODE);
    }
    return arg_opt.unwrap();
}
