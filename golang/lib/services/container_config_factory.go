package services

type ContainerConfigFactory interface {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetCreationConfig(containerIpAddr string) (*ContainerCreationConfig, error)

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*ContainerRunConfig, error)
}
