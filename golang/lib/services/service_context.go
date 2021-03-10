package services

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// This struct represents a Docker container running a service, and exposes functions for manipulating
// that container
type ServiceContext struct {
	client core_api_bindings.TestExecutionServiceClient
	serviceId ServiceID
	ipAddress string
}

func NewServiceContext(client core_api_bindings.TestExecutionServiceClient, serviceId ServiceID, ipAddress string) *ServiceContext {
	return &ServiceContext{client: client, serviceId: serviceId, ipAddress: ipAddress}
}

func (self *ServiceContext) GetServiceID() ServiceID {
	return self.serviceId
}

func (self *ServiceContext) GetIPAddress() string {
	return self.ipAddress
}

func (self *ServiceContext) ExecCommand(command []string) (int32, *[]byte, error) {
	serviceId := self.serviceId
	args := &core_api_bindings.ExecCommandArgs{
		ServiceId: string(serviceId),
		CommandArgs: command,
	}
	resp, err := self.client.ExecCommand(context.Background(), args)
	logrus.Infof("Exec Command Response: %+v", resp)
	if err != nil {
		return 0, nil, stacktrace.Propagate(
			err,
			"An error occurred executing command '%v' on service '%v'",
			command,
			serviceId)
	}
	return resp.ExitCode, &resp.LogOutput, nil
}
