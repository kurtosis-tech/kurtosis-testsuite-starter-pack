package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

const (
	port = 1323

	testVolumeMountpoint = "/test-volume"
)

type DatastoreContainerConfigFactory struct {
	dockerImage string
}

func NewDatastoreContainerConfigFactory(dockerImage string) *DatastoreContainerConfigFactory {
	return &DatastoreContainerConfigFactory{dockerImage: dockerImage}
}

func (factory DatastoreContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	serviceWrappingFunc := func(serviceCtx *services.ServiceContext) services.Service {
		return NewDatastoreService(serviceCtx, port)
	}
	result := services.NewContainerCreationConfigBuilder(
		factory.dockerImage,
		testVolumeMountpoint,
		serviceWrappingFunc,
	).WithUsedPorts(
		map[string]bool{
			fmt.Sprintf("%v/tcp", port): true,
		},
	).Build()
	return result, nil
}

func (factory DatastoreContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*services.ContainerRunConfig, error) {
	return services.NewContainerRunConfigBuilder().Build(), nil
}
