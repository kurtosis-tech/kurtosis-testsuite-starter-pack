package services

import "os"

// TODO TODO document all these
type ContainerConfigBuilder struct {
	image                    string
	testVolumeMountpoint     string
	usedPortsSet             map[string]bool
	serviceCreatingFunc      func(*ServiceContext) Service
	fileGeneratingFuncs      map[string]func(*os.File) error
	filesArtifactMountpoints map[FilesArtifactID]string
	entrypointOverrideFunc 	 func(ipAddr string, generatedFileFilepaths map[string]string)
	cmdOverrideFunc		     func(ipAddr string, generatedFileFilepaths map[string]string)
	envVarOverridesFunc      func(ipAddr string, generatedFileFilepaths map[string]string)
}

func NewContainerConfigBuilder(image string, testVolumeMountpoint string, serviceCreatingFunc func(ctx *ServiceContext) Service) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:                    image,
		testVolumeMountpoint:     testVolumeMountpoint,
		usedPortsSet:             map[string]bool{},
		serviceCreatingFunc:      serviceCreatingFunc,
		fileGeneratingFuncs:      map[string]func(file *os.File) error{},
		filesArtifactMountpoints: map[FilesArtifactID]string{},
	}
}

func (builder *ContainerConfigBuilder) WithUsedPorts(usedPortsSet map[string]bool) *ContainerConfigBuilder {
	builder.usedPortsSet = usedPortsSet
	return builder
}

func (builder *ContainerConfigBuilder) WithGeneratedFiles(fileGeneratingFuncs map[string]func(*os.File) error) *ContainerConfigBuilder {
	builder.fileGeneratingFuncs = fileGeneratingFuncs
	return builder
}

func (builder *ContainerConfigBuilder) WithFilesArtifacts(filesArtifactMountpoints map[FilesArtifactID]string) *ContainerConfigBuilder {
	builder.filesArtifactMountpoints = filesArtifactMountpoints
	return builder
}


func (builder *ContainerConfigBuilder) Build() *ContainerCreationConfig {
	return &ContainerCreationConfig{
		image:                        builder.image,
		testVolumeMountpoint:         builder.testVolumeMountpoint,
		usedPortsSet:                 builder.usedPortsSet,
		serviceCreatingFunc:          builder.serviceCreatingFunc,
		fileGeneratingFuncs:          builder.fileGeneratingFuncs,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
	}
}


