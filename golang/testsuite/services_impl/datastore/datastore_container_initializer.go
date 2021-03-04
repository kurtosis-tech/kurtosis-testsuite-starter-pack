/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"os"
)

const (
	port = 1323

	testVolumeMountpoint = "/test-volume"
)

type DatastoreContainerInitializer struct {
	dockerImage string
}

func NewDatastoreContainerInitializer(dockerImage string) *DatastoreContainerInitializer {
	return &DatastoreContainerInitializer{dockerImage: dockerImage}
}

func (d DatastoreContainerInitializer) GetDockerImage() string {
	return d.dockerImage
}

func (d DatastoreContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		fmt.Sprintf("%v/tcp", port): true,
	}
}

func (d DatastoreContainerInitializer) GetServiceWrappingFunc() func (ctx *services.ServiceContext) services.Service {
	return func(ctx *services.ServiceContext) services.Service {
		return NewDatastoreService(ctx, port);
	}
}

func (d DatastoreContainerInitializer) GetFilesToGenerate() map[string]bool {
	return map[string]bool{}
}

func (d DatastoreContainerInitializer) InitializeGeneratedFiles(filesToGenerate map[string]*os.File) error {
	return nil
}

func (d DatastoreContainerInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	return map[services.FilesArtifactID]string{}
}

func (d DatastoreContainerInitializer) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func (d DatastoreContainerInitializer) GetStartCommandOverrides(
		mountedFileFilepaths map[string]string,
		ipAddr string) (entrypointArgs []string, cmdArgs []string, resultErr error) {
	// We have a launch CMD specified in the Dockerfile the datastore service was built with and we don't need
	// to specify an ENTRYPOINT, so we leave everything nil
	return nil, nil, nil
}

