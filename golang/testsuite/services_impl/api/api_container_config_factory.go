package api

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	Port = 2434

	configFileKey = "config-file"

	testVolumeMountpoint = "/test-volume"
)

// Fields are public so we can marshal them as JSON
type config struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort int    `json:"datastorePort"`
}

type ApiContainerConfigFactory struct {
	image              string
	datastoreIPAddress string
	datastorePort      int
}

func NewApiContainerConfigFactory(image string, datastoreIPAddress string, datastorePort int) *ApiContainerConfigFactory {
	return &ApiContainerConfigFactory{image: image, datastoreIPAddress: datastoreIPAddress, datastorePort: datastorePort}
}

func (factory ApiContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	configInitializingFunc := func(fp *os.File) error {
		logrus.Debugf("Datastore IP: %v , port: %v", factory.datastoreIPAddress, factory.datastorePort)
		configObj := config{
			DatastoreIp:   factory.datastoreIPAddress,
			DatastorePort: factory.datastorePort,
		}
		configBytes, err := json.Marshal(configObj)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
		}

		logrus.Debugf("API config JSON: %v", string(configBytes))

		if _, err := fp.Write(configBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
		}

		return nil
	}

	result := services.NewContainerCreationConfigBuilder(
		factory.image,
		testVolumeMountpoint,
	).WithUsedPorts(map[string]bool{
		fmt.Sprintf("%v/tcp", Port): true,
	}).WithGeneratedFiles(map[string]func(*os.File) error{
		configFileKey: configInitializingFunc,
	}).Build()

	return result, nil
}

func (factory ApiContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*services.ContainerRunConfig, error) {
	configFilepath, found := generatedFileFilepaths[configFileKey]
	if !found {
		return nil, stacktrace.NewError("No filepath found for config file key '%v'", configFileKey)
	}
	startCmd := []string{
		"./api.bin",
		"--config",
		configFilepath,
	}
	result := services.NewContainerRunConfigBuilder().WithCmdOverride(startCmd).Build()
	return result, nil
}
