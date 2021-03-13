package services

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/palantir/stacktrace"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type ServiceContext struct {
	client core_api_bindings.TestExecutionServiceClient
	serviceId ServiceID
	ipAddress string
}

func NewServiceContext(client core_api_bindings.TestExecutionServiceClient, serviceId ServiceID, ipAddress string) *ServiceContext {
	return &ServiceContext{client: client, serviceId: serviceId, ipAddress: ipAddress}
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
