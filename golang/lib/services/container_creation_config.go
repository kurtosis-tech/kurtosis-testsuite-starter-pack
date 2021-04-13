package services

import "os"

// TODO defensive copy when we're giving back complex objects?????
type ContainerCreationConfig struct {
	image                        string
	testVolumeMountpoint         string
	usedPortsSet                 map[string]bool
	serviceCreatingFunc          func(*ServiceContext) Service
	fileGeneratingFuncs          map[string]func(*os.File) error
	filesArtifactMountpoints     map[FilesArtifactID]string
}

func (config *ContainerCreationConfig) GetImage() string {
	return config.image
}

func (config *ContainerCreationConfig) GetTestVolumeMountpoint() string {
	return config.testVolumeMountpoint
}

func (config *ContainerCreationConfig) GetUsedPortsSet() map[string]bool {
	return config.usedPortsSet
}

func (config *ContainerCreationConfig) GetServiceCreatingFunc() func(ctx *ServiceContext) Service {
	return config.serviceCreatingFunc
}

func (config *ContainerCreationConfig) GetFileGeneratingFuncs() map[string]func(*os.File) error {
	return config.fileGeneratingFuncs
}

func (config *ContainerCreationConfig) GetFilesArtifactMountpoints() map[FilesArtifactID]string {
	return config.filesArtifactMountpoints
}
