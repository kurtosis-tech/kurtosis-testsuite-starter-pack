package nginx_static

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"strconv"
)

const (
	dockerImage = "flashspys/nginx-static"

	listenPort = 80

	testVolumeMountpoint = "/test-volume"

	nginxStaticFilesDirpath = "/static"
)

/*
A config factory implementation to launch an NginxStaticService pre-initialized with the contents of
	the given files artifact
*/
type NginxStaticContainerConfigFactory struct {
	filesArtifactIdOpt services.FilesArtifactID
}

// NOTE: The files artifact ID is optional; if it's emptystring then no files artifact will be extracted
func NewNginxStaticContainerConfigFactory(filesArtifactIdOpt services.FilesArtifactID) *NginxStaticContainerConfigFactory {
	return &NginxStaticContainerConfigFactory{filesArtifactIdOpt: filesArtifactIdOpt}
}

func (factory NginxStaticContainerConfigFactory) GetCreationConfig(containerIpAddr string) (*services.ContainerCreationConfig, error) {
	builder := services.NewContainerCreationConfigBuilder(
		dockerImage,
		testVolumeMountpoint,
		func(serviceCtx *services.ServiceContext) services.Service { return NewNginxStaticService(serviceCtx) },
	).WithUsedPorts(map[string]bool{
		strconv.Itoa(listenPort): true,
	})

	if factory.filesArtifactIdOpt != "" {
		builder.WithFilesArtifacts(map[services.FilesArtifactID]string{
			factory.filesArtifactIdOpt: nginxStaticFilesDirpath,
		})
	}

	return builder.Build(), nil
}

func (factory NginxStaticContainerConfigFactory) GetRunConfig(containerIpAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	return services.NewContainerRunConfigBuilder().Build(), nil
}

func (factory NginxStaticContainerConfigFactory) GetPort() int {
	return listenPort
}
