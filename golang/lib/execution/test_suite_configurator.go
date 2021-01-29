/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution

import "github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"

// Implementations of this interface are responsible for initialzing the testsuite to a state
//  where it can be run
type TestSuiteConfigurator interface {
	/*
	This function should be used to configure the testsuite's logging framework, and will be run
		before the testsuite is run

	Args:
		logLevelStr: The testsuite log level string passed in at runtime, which should be parsed
			 so that the logging framework can be configured.

	 */
	SetLogLevel(logLevelStr string) error

	/*
	This function should parse the custom testsuite parameters JSON (if any) and create an instance
		of the testsuite.

	Args:
		paramsJsonStr: The JSON-serialized custom params data used for configuring testsuite behaviour
			that was passed in when Kurtosis was started.
	 */
	ParseParamsAndCreateSuite(paramsJsonStr string) (testsuite.TestSuite, error)
}
