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
	entrypointOverride       []string
	cmdOverride              []string
	environmentVarOverrides  map[string]string
}

func NewContainerConfigBuilder(image string, testVolumeMountpoint string, serviceCreatingFunc func(ctx *ServiceContext) Service) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:                    image,
		testVolumeMountpoint:     testVolumeMountpoint,
		usedPortsSet:             map[string]bool{},
		serviceCreatingFunc:      serviceCreatingFunc,
		fileGeneratingFuncs:      map[string]func(file *os.File) error{},
		filesArtifactMountpoints: map[FilesArtifactID]string{},
		entrypointOverride:       nil,
		cmdOverride:              nil,
		environmentVarOverrides:  map[string]string{},
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

func (builder *ContainerConfigBuilder) WithEntrypointOverride(args []string) *ContainerConfigBuilder {
	builder.entrypointOverride = args
	return builder
}

func (builder *ContainerConfigBuilder) WithCmdOverride(args []string) *ContainerConfigBuilder {
	builder.cmdOverride = args
	return builder
}

func (builder *ContainerConfigBuilder) WithEnvironmentVariableOverrides(envVars map[string]string) *ContainerConfigBuilder {
	builder.environmentVarOverrides = envVars
	return builder
}

func (builder *ContainerConfigBuilder) Build() *ContainerConfig {
	return &ContainerConfig{
		image:                        builder.image,
		testVolumeMountpoint:         builder.testVolumeMountpoint,
		usedPortsSet:                 builder.usedPortsSet,
		serviceCreatingFunc:          builder.serviceCreatingFunc,
		fileGeneratingFuncs:          builder.fileGeneratingFuncs,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
		entrypointOverrideArgs:       builder.entrypointOverride,
		cmdOverrideArgs:              builder.cmdOverride,
		environmentVariableOverrides: builder.environmentVarOverrides,
	}
}


