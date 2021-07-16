/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_env_vars

/*
A package to contain the contract of Docker environment variables that the testsuite container accepts
 */

type TestSuiteMode string
const (
	CustomParamsJsonEnvVar  = "CUSTOM_PARAMS_JSON"
	DebuggerPortEnvVar      = "DEBUGGER_PORT"
	KurtosisApiSocketEnvVar = "KURTOSIS_API_SOCKET" // Only populated if in test-running mode
	LogLevelEnvVar          = "LOG_LEVEL"
)

