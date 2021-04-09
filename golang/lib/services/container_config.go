package services

import "os"

// TODO defensive copy when we're giving back complex objects?????
type ContainerConfig struct {
	image                        string
	testVolumeMountpoint         string
	usedPortsSet                 map[string]bool
	serviceCreatingFunc          func(*ServiceContext) Service
	fileGeneratingFuncs          map[string]func(*os.File) error
	filesArtifactMountpoints     map[FilesArtifactID]string
	entrypointOverrideArgs       []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
}

func (config *ContainerConfig) GetImage() string {
	return config.image
}

func (config *ContainerConfig) GetTestVolumeMountpoint() string {
	return config.testVolumeMountpoint
}

func (config *ContainerConfig) GetUsedPortsSet() map[string]bool {
	return config.usedPortsSet
}

func (config *ContainerConfig) GetServiceCreatingFunc() func(ctx *ServiceContext) Service {
	return config.serviceCreatingFunc
}

func (config *ContainerConfig) GetFileGeneratingFuncs() map[string]func(*os.File) error {
	return config.fileGeneratingFuncs
}

func (config *ContainerConfig) GetFilesArtifactMountpoints() map[FilesArtifactID]string {
	return config.filesArtifactMountpoints
}

func (config *ContainerConfig) GetEntrypointOverrideArgs() []string {
	return config.entrypointOverrideArgs
}

func (config *ContainerConfig) GetCmdOverrideArgs() []string {
	return config.cmdOverrideArgs
}

func (config *ContainerConfig) GetEnvironmentVariableOverrides() map[string]string {
	return config.environmentVariableOverrides
}