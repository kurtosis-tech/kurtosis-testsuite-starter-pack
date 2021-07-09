package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

const (
	Port = 1323

	testVolumeMountpoint = "/test-volume"
)

type DatastoreContainerConfigFactory struct {
	dockerImage string
}

func NewDatastoreContainerConfigFactory(dockerImage string) *DatastoreContainerConfigFactory {
	return &DatastoreContainerConfigFactory{dockerImage: dockerImage}
}

func (factory DatastoreContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	result := services.NewContainerCreationConfigBuilder(
		factory.dockerImage,
		testVolumeMountpoint,
	).WithUsedPorts(
		map[string]bool{
			fmt.Sprintf("%v/tcp", Port): true,
		},
	).Build()
	return result, nil
}

func (factory DatastoreContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	return services.NewContainerRunConfigBuilder().Build(), nil
}
