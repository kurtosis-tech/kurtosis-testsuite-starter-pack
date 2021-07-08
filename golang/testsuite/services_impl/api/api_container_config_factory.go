package api

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	port = 2434

	configFileKey = "config-file"

	testVolumeMountpoint = "/test-volume"
)

// Fields are public so we can marshal them as JSON
type config struct {
	DatastoreIp string	`json:"datastoreIp"`
	DatastorePort int	`json:"datastorePort"`
}

type ApiContainerConfigFactory struct {
	image     string
	datastore *datastore.DatastoreService
}

func NewApiContainerConfigFactory(image string, datastore *datastore.DatastoreService) *ApiContainerConfigFactory {
	return &ApiContainerConfigFactory{image: image, datastore: datastore}
}


func (factory ApiContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	configInitializingFunc := func(fp *os.File) error {
		logrus.Debugf("Datastore IP: %v , port: %v", factory.datastore.GetServiceContext().GetIPAddress(), factory.datastore.GetPort())
		configObj := config{
			DatastoreIp:   factory.datastore.GetServiceContext().GetIPAddress(),
			DatastorePort: factory.datastore.GetPort(),
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
		func(serviceCtx *services.ServiceContext) services.Service { return NewApiService(serviceCtx, port) },
	).WithUsedPorts(map[string]bool{
		fmt.Sprintf("%v/tcp", port): true,
	}).WithGeneratedFiles(map[string]func(*os.File) error{
		configFileKey: configInitializingFunc,
	}).Build()

	return result, nil
}

func (factory ApiContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
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