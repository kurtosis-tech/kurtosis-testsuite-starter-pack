/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type TestContext struct {}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (context TestContext) Fatal(err error) {
	// We rely on panicking here because we want to completely stop whatever the test is doing
	failTest(err)
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (context TestContext) AssertTrue(condition bool, err error) {
	if (!condition) {
		failTest(err)
	}
}

func failTest(err error) {
	panic(err)
}
