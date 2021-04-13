package services

type ContainerConfigFactory interface {
	// TODO DOCS LINK
	GetCreationConfig(containerIpAddr string) (*ContainerCreationConfig, error)

	// TODO DOCS Link
	GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string) (*ContainerRunConfig, error)
}
