package my_custom_service
/*
	NEW USER ONBOARDING:
	- Rename this package, this file, and the containing directory after your custom service.
	- Rename all structs and functions in this file after your custom service.
*/

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
)

const (
	testVolumeMountpoint = "/test-volume"
	/*
		NEW USER ONBOARDING:
		- Change this port number to the primary communication port that your service uses to speak to other services, or to clients.
	*/
	port = 1234
)

type MyCustomServiceConfigFactory struct {
	image     string
	port	  int
}

func NewMyCustomServiceContainerConfigFactory(image string) *MyCustomServiceConfigFactory {
	return &MyCustomServiceConfigFactory{image: image}
}


func (factory MyCustomServiceConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	result := services.NewContainerCreationConfigBuilder(
		factory.image,
		testVolumeMountpoint,
		func(serviceCtx *services.ServiceContext) services.Service { return NewMyCustomService(serviceCtx, port) },
	).WithUsedPorts(map[string]bool{
		/*
			NEW USER ONBOARDING:
			- Add any other ports that your service needs to have open to other services, or to the
		*/
		fmt.Sprintf("%v/tcp", port): true,
	}).Build()

	return result, nil
}

func (factory MyCustomServiceConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*services.ContainerRunConfig, error) {
	/*
		NEW USER ONBOARDING:
		- Change this start command to reflect the actual start command of your custom service.
	 */
	startCmd := []string{
		"/path/to/start/binary",
		"--<config-flag-1",
		"config-flag-value-1",
		"--<config-flag-2",
		"config-flag-value-2",
		"...",
	}
	result := services.NewContainerRunConfigBuilder().WithCmdOverride(startCmd).Build()
	return result, nil
}
