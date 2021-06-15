package my_custom_test
/*
	NEW USER ONBOARDING:
	- Rename this package, this file, and the containing directory to reflect the functionality of your custom test.
*/
import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/sirupsen/logrus"
)

type MyCustomTest struct {}

func (e MyCustomTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(30).WithRunTimeoutSeconds(30)
}

func (e MyCustomTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	logrus.Infof("Setting up custom test.")
	/*
		NEW USER ONBOARDING:
		- Fill in the logic necessary to set up your custom testnet.
	*/
	return nil, nil
}

func (e MyCustomTest) Run(uncastedNetwork networks.Network) error {
	logrus.Infof("Running custom test.")
	/*
		NEW USER ONBOARDING:
		- Fill in the logic necessary to run your custom test.
	*/
	return nil
}