package services

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/palantir/stacktrace"
	"path"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type GeneratedFileFilepaths struct {
	AbsoluteFilepathOnTestsuiteContainer string
	AbsoluteFilepathOnServiceContainer   string
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type ServiceContext struct {
	client                                   core_api_bindings.ApiContainerServiceClient
	serviceId                                ServiceID
	ipAddress                                string
	testVolumeMountpointOnTestsuiteContainer string
	testVolumeMountpointOnServiceContainer   string
}

func NewServiceContext(
		client core_api_bindings.ApiContainerServiceClient,
		serviceId ServiceID,
		ipAddress string,
		testVolumeMountpointOnTestsuiteContainer string,
		testVolumeMountpointOnServiceContainer string) *ServiceContext {
	return &ServiceContext{
		client:                                   client,
		serviceId:                                serviceId,
		ipAddress:                                ipAddress,
		testVolumeMountpointOnTestsuiteContainer: testVolumeMountpointOnTestsuiteContainer,
		testVolumeMountpointOnServiceContainer: testVolumeMountpointOnServiceContainer,
	}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (self *ServiceContext) GetServiceID() ServiceID {
	return self.serviceId
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (self *ServiceContext) GetIPAddress() string {
	return self.ipAddress
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (self *ServiceContext) ExecCommand(command []string) (int32, *[]byte, error) {
	serviceId := self.serviceId
	args := &core_api_bindings.ExecCommandArgs{
		ServiceId: string(serviceId),
		CommandArgs: command,
	}
	resp, err := self.client.ExecCommand(context.Background(), args)
	if err != nil {
		return 0, nil, stacktrace.Propagate(
			err,
			"An error occurred executing command '%v' on service '%v'",
			command,
			serviceId)
	}
	return resp.ExitCode, &resp.LogOutput, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
func (self *ServiceContext) GenerateFiles(filesToGenerateSet map[string]bool) (map[string]*GeneratedFileFilepaths, error) {
	serviceId := self.serviceId
	fileGenerationOpts := map[string]*core_api_bindings.FileGenerationOptions{}
	for fileId := range filesToGenerateSet {
		fileGenerationOpts[fileId] = &core_api_bindings.FileGenerationOptions{
			FileTypeToGenerate: core_api_bindings.FileGenerationOptions_FILE,
		}
	}
	args := &core_api_bindings.GenerateFilesArgs{
		ServiceId:       string(serviceId),
		FilesToGenerate: fileGenerationOpts,
	}
	resp, err := self.client.GenerateFiles(context.Background(), args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating files using args: %+v", args)
	}
	generatedFileRelativeFilepaths := resp.GeneratedFileRelativeFilepaths

	result := map[string]*GeneratedFileFilepaths{}
	for fileId := range filesToGenerateSet {
		relativeFilepath, found := generatedFileRelativeFilepaths[fileId]
		if !found {
			return nil, stacktrace.NewError(
				"No filepath (relative to test volume root) was returned for file '%v', even though we requested it; this is a Kurtosis bug",
				fileId)
		}
		absFilepathOnTestsuite := path.Join(self.testVolumeMountpointOnTestsuiteContainer, relativeFilepath)
		absFilepathOnService := path.Join(self.testVolumeMountpointOnServiceContainer, relativeFilepath)
		result[fileId] = &GeneratedFileFilepaths{
			AbsoluteFilepathOnTestsuiteContainer: absFilepathOnTestsuite,
			AbsoluteFilepathOnServiceContainer:   absFilepathOnService,
		}
	}
	return result, nil
}
