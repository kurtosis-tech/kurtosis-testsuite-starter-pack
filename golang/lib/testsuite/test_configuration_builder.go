package testsuite

import "github.com/kurtosis-tech/kurtosis-client/golang/lib/services"

const (
	// vvvvvvvvv Update the docs if you change these vvvvvvvvvvv
	defaultSetupTimeoutSeconds = 180;
	defaultRunTimeoutSeconds = 180;
	defaultPartitioningEnabled = false;
	// ^^^^^^^^^ Update the docs if you change these ^^^^^^^^^^^
)

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type TestConfigurationBuilder struct {
	setupTimeoutSeconds uint32
	runTimeoutSeconds uint32
	isPartioningEnabled bool
	filesArtifactUrls map[services.FilesArtifactID]string
}

func NewTestConfigurationBuilder() *TestConfigurationBuilder {
	return &TestConfigurationBuilder{
		setupTimeoutSeconds: defaultSetupTimeoutSeconds,
		runTimeoutSeconds:   defaultRunTimeoutSeconds,
		isPartioningEnabled: defaultPartitioningEnabled,
		filesArtifactUrls:   map[services.FilesArtifactID]string{},
	}
}

func (builder *TestConfigurationBuilder) WithSetupTimeoutSeconds(setupTimeoutSeconds uint32) *TestConfigurationBuilder {
	builder.setupTimeoutSeconds = setupTimeoutSeconds
	return builder
}

func (builder *TestConfigurationBuilder) WithRunTimeoutSeconds(runTimeoutSeconds uint32) *TestConfigurationBuilder {
	builder.runTimeoutSeconds = runTimeoutSeconds
	return builder
}

func (builder *TestConfigurationBuilder) WithPartitioningEnabled(isPartitioningEnabled bool) *TestConfigurationBuilder {
	builder.isPartioningEnabled = isPartitioningEnabled
	return builder
}

func (builder *TestConfigurationBuilder) WithFilesArtifactUrls(filesArtifactUrls map[services.FilesArtifactID]string) *TestConfigurationBuilder {
	builder.filesArtifactUrls = filesArtifactUrls
	return builder
}

func (builder TestConfigurationBuilder) Build() *TestConfiguration {
	return &TestConfiguration{
		SetupTimeoutSeconds:   builder.setupTimeoutSeconds,
		RunTimeoutSeconds:     builder.runTimeoutSeconds,
		IsPartitioningEnabled: builder.isPartioningEnabled,
		FilesArtifactUrls:     builder.filesArtifactUrls,
	}
}
