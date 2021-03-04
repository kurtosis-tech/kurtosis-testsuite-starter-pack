package exec_cmd_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"os"
)

const (
	testVolumeMountpoint = "/test-volume"
)

type ExecCmdTestContainerInitializer struct {
	dockerImage string
}

func NewExecCmdTestContainerInitializer(dockerImage string) *ExecCmdTestContainerInitializer {
	return &ExecCmdTestContainerInitializer{dockerImage: dockerImage}
}

func (self ExecCmdTestContainerInitializer) GetDockerImage() string {
	return self.dockerImage
}

func (self ExecCmdTestContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{}
}

func (self ExecCmdTestContainerInitializer) GetServiceWrappingFunc() func(ctx *services.ServiceContext) services.Service {
	return func(ctx *services.ServiceContext) services.Service {
		return NewExecCmdTestService(ctx)
	}
}

func (self ExecCmdTestContainerInitializer) GetFilesToGenerate() map[string]bool {
	return map[string]bool{}
}

func (self ExecCmdTestContainerInitializer) InitializeGeneratedFiles(generatedFiles map[string]*os.File) error {
	// No generated files
	return nil
}

func (self ExecCmdTestContainerInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	// No files artifacts
	return map[services.FilesArtifactID]string{}
}

func (self ExecCmdTestContainerInitializer) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func (self ExecCmdTestContainerInitializer) GetStartCommandOverrides(
		generatedFileFilepaths map[string]string,
		ipAddr string) (entrypointArgs []string, cmdArgs []string, resultErr error) {
	// We sleep because the only function of this container is to test Docker exec'ing a command while it's running
	// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
	// args), but this provides a nice little regression test of the ENTRYPOINT overriding
	return []string{"sleep"}, []string{"30"}, nil
}

