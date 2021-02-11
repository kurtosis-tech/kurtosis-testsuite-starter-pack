/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
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

type ApiContainerInitializer struct {
	dockerImage string
	datastore *datastore.DatastoreService
}

func NewApiContainerInitializer(dockerImage string, datastore *datastore.DatastoreService) *ApiContainerInitializer {
	return &ApiContainerInitializer{dockerImage: dockerImage, datastore: datastore}
}

func (initializer ApiContainerInitializer) GetDockerImage() string {
	return initializer.dockerImage
}

func (initializer ApiContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		fmt.Sprintf("%v/tcp", port): true,
	}
}

func (initializer ApiContainerInitializer) GetServiceWrappingFunc() func(serviceId services.ServiceID, ipAddr string) services.Service {
	return func(serviceId services.ServiceID, ipAddr string) services.Service {
		return NewApiService(serviceId, ipAddr, port);
	};
}

func (initializer ApiContainerInitializer) GetFilesToMount() map[string]bool {
	return map[string]bool{
		configFileKey: true,
	}
}

func (initializer ApiContainerInitializer) InitializeMountedFiles(mountedFiles map[string]*os.File) error {
	logrus.Debugf("Datastore IP: %v , port: %v", initializer.datastore.GetIPAddress(), initializer.datastore.GetPort())
	configObj := config{
		DatastoreIp:   initializer.datastore.GetIPAddress(),
		DatastorePort: initializer.datastore.GetPort(),
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	logrus.Debugf("API config JSON: %v", string(configBytes))

	configFp := mountedFiles[configFileKey]
	if _, err := configFp.Write(configBytes); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return nil
}

func (initializer ApiContainerInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	return map[services.FilesArtifactID]string{}
}

func (initializer ApiContainerInitializer) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func (initializer ApiContainerInitializer) GetStartCommand(mountedFileFilepaths map[string]string, ipPlaceholder string) ([]string, error) {
	// TODO Replace this with a productized way to start a container using only environment variables
	configFilepath := mountedFileFilepaths[configFileKey]
	startCmd := []string{
		"./api.bin",
		"--config",
		configFilepath,
	}
	return startCmd, nil
}

