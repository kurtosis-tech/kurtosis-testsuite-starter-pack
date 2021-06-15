package exec_cmd_test

import "github.com/kurtosis-tech/kurtosis-client/golang/services"

const (
	testVolumeMountpoint = "/test-volume"
)

type ExecCmdTestContainerConfigFactory struct {
	image string
}

func NewExecCmdTestContainerConfigFactory(image string) *ExecCmdTestContainerConfigFactory {
	return &ExecCmdTestContainerConfigFactory{image: image}
}

func (e ExecCmdTestContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	result := services.NewContainerCreationConfigBuilder(
		e.image,
		testVolumeMountpoint,
		func(serviceCtx *services.ServiceContext) services.Service { return NewExecCmdTestService(serviceCtx) },
	).Build()
	return result, nil
}

func (e ExecCmdTestContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*services.ContainerRunConfig, error) {
	// We sleep because the only function of this container is to test Docker exec'ing a command while it's running
	// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
	// args), but this provides a nice little regression test of the ENTRYPOINT overriding
	entrypointArgs := []string{"sleep"}
	cmdArgs := []string{"30"}
	result := services.NewContainerRunConfigBuilder().WithEntrypointOverride(
		entrypointArgs,
	).WithCmdOverride(
		cmdArgs,
	).Build()
	return result, nil
}

