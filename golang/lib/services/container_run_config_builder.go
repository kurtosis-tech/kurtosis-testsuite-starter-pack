package services

// TODO Defensive copies on all these With... functions???
type ContainerRunConfigBuilder struct {
	entrypointOverrideArgs  []string
	cmdOverrideArgs         []string
	environmentVariableOverrides     map[string]string
}

func NewContainerRunConfigBuilder() *ContainerRunConfigBuilder {
	return &ContainerRunConfigBuilder{
		entrypointOverrideArgs: nil,
		cmdOverrideArgs: nil,
		environmentVariableOverrides: map[string]string{},
	}
}

func (builder *ContainerRunConfigBuilder) WithEntrypointOverride(args []string) *ContainerRunConfigBuilder {
	builder.entrypointOverrideArgs = args
	return builder
}

func (builder *ContainerRunConfigBuilder) WithCmdOverride(args []string) *ContainerRunConfigBuilder {
	builder.cmdOverrideArgs = args
	return builder
}

func (builder *ContainerRunConfigBuilder) WithEnvironmentVariableOverrides(envVars map[string]string) *ContainerRunConfigBuilder {
	builder.environmentVariableOverrides = envVars
	return builder
}

func (builder *ContainerRunConfigBuilder) Build() *ContainerRunConfig {
	return &ContainerRunConfig{
		entrypointOverrideArgs:       builder.entrypointOverrideArgs,
		cmdOverrideArgs:              builder.cmdOverrideArgs,
		environmentVariableOverrides: builder.environmentVariableOverrides,
	}
}
