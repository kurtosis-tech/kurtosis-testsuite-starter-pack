/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

import (
	"fmt"
	"os"
)

type MockDockerContainerInitializer struct{}

func NewMockDockerContainerInitializer() *MockDockerContainerInitializer {
	return &MockDockerContainerInitializer{}
}

func (m MockDockerContainerInitializer) GetDockerImage() string {
	return "some-image"
}

func (m MockDockerContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		fmt.Sprintf("%v/tcp", MockServicePort): true,
	}
}

func (m MockDockerContainerInitializer) GetService(serviceId ServiceID, ipAddr string) Service {
	return NewMockService(serviceId, ipAddr, 1)
}

func (m MockDockerContainerInitializer) GetFilesToMount() map[string]bool {
	return map[string]bool{}
}

func (m MockDockerContainerInitializer) InitializeMountedFiles(mountedFiles map[string]*os.File) error {
	return nil
}

func (m MockDockerContainerInitializer) GetFilesArtifactMountpoints() map[FilesArtifactID]string {
	return map[FilesArtifactID]string{}
}

func (m MockDockerContainerInitializer) GetTestVolumeMountpoint() string {
	return "/test-volume"
}

func (m MockDockerContainerInitializer) GetStartCommand(mountedFileFilepaths map[string]string, ipPlaceholder string) ([]string, error) {
	return []string{
		"some-binary",
		"--some-flag",
	}, nil
}

