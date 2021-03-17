/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

import (
	"os"
)

// The ID of an artifact containing files that should be mounted into a service container
type FilesArtifactID string

// TODO Create a DockerContainerInitializerBuilder rather than forcing users to update their code with a new
//  method every time a new feature comes out!
// GENERIC TOOD: If Go had generics, this would be parameterized with the subtype of Service that this returns
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type DockerContainerInitializer interface {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetDockerImage() string

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetUsedPorts() map[string]bool

	// GENERICS TOOD: When Go has generics, make this return type be parameterized
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetService(serviceCtx *ServiceContext) Service

	// GENERICS TOOD: If Go had generics, we could parameterize this entire class with an enum of the types of files this service consumes
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetFilesToGenerate() map[string]bool

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	InitializeGeneratedFiles(generatedFiles map[string]*os.File) error

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetFilesArtifactMountpoints() map[FilesArtifactID]string

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetTestVolumeMountpoint() string

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetStartCommandOverrides(generatedFileFilepaths map[string]string, ipAddr string) (entrypointArgs []string, cmdArgs []string, err error)

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetEnvironmentVariableOverrides() (map[string]string, error)
}
