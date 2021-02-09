use std::error::Error;
use anyhow::Result;

// Implementations of this interface are responsible for initialzing the testsuite to a state
//  where it can be run
pub trait TestSuiteConfigurator {
	/*
	This function should be used to configure the testsuite's logging framework, and will be run
		before the testsuite is run

	Args:
		log_level_str: The testsuite log level string passed in at runtime, which should be parsed
			 so that the logging framework can be configured.

	 */
	fn set_log_level(&self, log_level_str: &str) -> Result<()>;

	/*
	This function should parse the custom testsuite parameters JSON (if any) and create an instance
		of the testsuite.

	Args:
		params_json_str: The JSON-serialized custom params data used for configuring testsuite behaviour
			that was passed in when Kurtosis was started.
	 */
	// TODO MAKE RETURN TYPE OF TYPE TESTSUITE
	fn parse_params_and_create_suite(&self, params_json_str: &str) -> Result<()>;
}
