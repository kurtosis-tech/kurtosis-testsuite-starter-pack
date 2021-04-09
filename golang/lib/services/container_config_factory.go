package services

type ContainerConfigFactory interface {
	Create(containerIpAddr string) *ContainerConfig
}
