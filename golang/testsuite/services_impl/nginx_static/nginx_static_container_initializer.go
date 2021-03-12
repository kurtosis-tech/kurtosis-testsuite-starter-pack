/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package nginx_static

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"os"
	"strconv"
)

/*
A DockerContainerInitializer to launch an NginxStaticService pre-initialized with the contents of
	the given files artifact
 */
type NginxStaticContainerInitializer struct {
	filesArtifactId services.FilesArtifactID
}

func NewNginxStaticContainerInitializer(filesArtifactId services.FilesArtifactID) *NginxStaticContainerInitializer {
	return &NginxStaticContainerInitializer{filesArtifactId: filesArtifactId}
}

func (s NginxStaticContainerInitializer) GetDockerImage() string {
	return dockerImage
}

func (s NginxStaticContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		strconv.Itoa(listenPort): true,
	}
}

func (s NginxStaticContainerInitializer) GetService(serviceCtx *services.ServiceContext) services.Service {
	return NewNginxStaticService(serviceCtx)
}

func (s NginxStaticContainerInitializer) GetFilesToGenerate() map[string]bool {
	// No generated files to mount
	return map[string]bool{}
}

func (s NginxStaticContainerInitializer) InitializeGeneratedFiles(filesToGenerate map[string]*os.File) error {
	// No generated files to initialize
	return nil
}

func (s NginxStaticContainerInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	return map[services.FilesArtifactID]string{
		s.filesArtifactId: nginxStaticFilesDirpath,
	}
}

func (s NginxStaticContainerInitializer) GetTestVolumeMountpoint() string {
	return "/test-volume"
}

func (s NginxStaticContainerInitializer) GetStartCommandOverrides(
		mountedFileFilepaths map[string]string,
		ipAddr string) (entrypointArgs []string, cmdArgs []string, resultErr error) {
	// Don't need any ENTRYPOINT or CMD overrides - use the default ones
	return nil, nil, nil
}
