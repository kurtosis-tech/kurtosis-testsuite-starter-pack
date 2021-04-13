package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
)

const (
	port = 1323

	testVolumeMountpoint = "/test-volume"
)

type DatastoreContainerConfigFactory struct {
	dockerImage string
}

func (factory DatastoreContainerConfigFactory) Create(containerIpAddr string) *services.ContainerCreationConfig {
	serviceWrappingFunc := func(serviceCtx *services.ServiceContext) services.Service {
		return NewDatastoreService(serviceCtx, port)
	}
	return services.NewContainerConfigBuilder(
		factory.dockerImage,
		testVolumeMountpoint,
		serviceWrappingFunc,
	).WithUsedPorts(
		map[string]bool{
			fmt.Sprintf("%v/tcp", port): true,
		},
	).Build()
}