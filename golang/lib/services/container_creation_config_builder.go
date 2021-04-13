package services

import "os"

// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type ContainerCreationConfigBuilder struct {
	image                    string
	testVolumeMountpoint     string
	usedPortsSet             map[string]bool
	serviceCreatingFunc      func(*ServiceContext) Service
	fileGeneratingFuncs      map[string]func(*os.File) error
	filesArtifactMountpoints map[FilesArtifactID]string
}

func NewContainerCreationConfigBuilder(image string, testVolumeMountpoint string, serviceCreatingFunc func(ctx *ServiceContext) Service) *ContainerCreationConfigBuilder {
	return &ContainerCreationConfigBuilder{
		image:                    image,
		testVolumeMountpoint:     testVolumeMountpoint,
		usedPortsSet:             map[string]bool{},
		serviceCreatingFunc:      serviceCreatingFunc,
		fileGeneratingFuncs:      map[string]func(file *os.File) error{},
		filesArtifactMountpoints: map[FilesArtifactID]string{},
	}
}

func (builder *ContainerCreationConfigBuilder) WithUsedPorts(usedPortsSet map[string]bool) *ContainerCreationConfigBuilder {
	builder.usedPortsSet = usedPortsSet
	return builder
}

func (builder *ContainerCreationConfigBuilder) WithGeneratedFiles(fileGeneratingFuncs map[string]func(*os.File) error) *ContainerCreationConfigBuilder {
	builder.fileGeneratingFuncs = fileGeneratingFuncs
	return builder
}

func (builder *ContainerCreationConfigBuilder) WithFilesArtifacts(filesArtifactMountpoints map[FilesArtifactID]string) *ContainerCreationConfigBuilder {
	builder.filesArtifactMountpoints = filesArtifactMountpoints
	return builder
}


func (builder *ContainerCreationConfigBuilder) Build() *ContainerCreationConfig {
	return &ContainerCreationConfig{
		image:                        builder.image,
		testVolumeMountpoint:         builder.testVolumeMountpoint,
		usedPortsSet:                 builder.usedPortsSet,
		serviceCreatingFunc:          builder.serviceCreatingFunc,
		fileGeneratingFuncs:          builder.fileGeneratingFuncs,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
	}
}


