/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/testsuite_impl/my_custom_test"
)

/*
	NEW USER ONBOARDING:
	- Refactor the name of the myCustomServiceImage property to reflect the name of your service.
*/
type ExampleTestsuite struct {
	myCustomServiceImage string
}

/*
	NEW USER ONBOARDING:
	- Refactor the name of the myCustomServiceImage argument to reflect the name of your service.
*/
func NewExampleTestsuite(myCustomServiceImage string) *ExampleTestsuite {
	return &ExampleTestsuite{myCustomServiceImage: myCustomServiceImage,}
}

func (suite ExampleTestsuite) GetTests() map[string]testsuite.Test {
	tests := map[string]testsuite.Test{
		"myCustomServiceTest": my_custom_test.NewMyCustomTest(suite.myCustomServiceImage),
	}

	return tests
}

func (suite ExampleTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}


