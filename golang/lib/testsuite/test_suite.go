/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite

/*
An interface which the user implements to register their tests.
 */
type TestSuite interface {
	// Get all the tests in the test suite; this is where users will "register" their tests
	GetTests() map[string]Test

	// Determines how many IP addresses will be available in the Docker network created for each test, which determines
	//  the maximum number of services that can be created in the test. The maximum number of services that each
	//  test can have = 2 ^ network_width_bits
	GetNetworkWidthBits() uint32
}
