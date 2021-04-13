package services

type ContainerRunConfig struct {
	entrypointOverrideArgs  []string
	cmdOverrideArgs         []string
	environmentVariableOverrides     map[string]string
}

func (config *ContainerRunConfig) GetEntrypointOverrideArgs() []string {
	return config.entrypointOverrideArgs
}

func (config *ContainerRunConfig) GetCmdOverrideArgs() []string {
	return config.cmdOverrideArgs
}

func (config *ContainerRunConfig) GetEnvironmentVariableOverrides() map[string]string {
	return config.environmentVariableOverrides
}
