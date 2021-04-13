package services


// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
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



// ====================================================================================================
//                                      Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
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
