package my_custom_test
/*
	NEW USER ONBOARDING:
	- Rename this package, this file, and the containing directory to reflect the functionality of your custom test.
	- Rename constants, structs, properties, and variables to reflect your new service name.
*/
import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/my_custom_service"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	myCustomServiceID services.ServiceID = "myCustomService"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls = 15
)

type MyCustomTest struct {
	MyCustomServiceImage string
}

func NewMyCustomTest(image string) *MyCustomTest {
	return &MyCustomTest{MyCustomServiceImage: image}
}

func (test MyCustomTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(30).WithRunTimeoutSeconds(30)
}

func (test MyCustomTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	logrus.Infof("Setting up custom test.")
	/*
		NEW USER ONBOARDING:
		- Add services multiple times using the below logic in order to have more than one service.
	*/
	configFactory := my_custom_service.NewMyCustomServiceContainerConfigFactory(test.MyCustomServiceImage)
	_, hostPortBindings, availabilityChecker, err := networkCtx.AddService(myCustomServiceID, configFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the service")
	}
	if err := availabilityChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the service to become available")
	}
	logrus.Infof("Added service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test MyCustomTest) Run(uncastedNetwork networks.Network) error {
	logrus.Infof("Running custom test.")
	// Necessary because Go doesn't have generics
	castedNetwork := uncastedNetwork.(*networks.NetworkContext)

	uncastedService, err := castedNetwork.GetService(myCustomServiceID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service")
	}

	// Necessary again due to no Go generics
	castedService := uncastedService.(*my_custom_service.MyCustomService)
	logrus.Infof("Service is available: %v", castedService.IsAvailable())

	/*
		NEW USER ONBOARDING:
		- Fill in the logic necessary to run your custom test.
	*/
	return nil
}