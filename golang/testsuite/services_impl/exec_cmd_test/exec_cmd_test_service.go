package exec_cmd_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

type ExecCmdTestService struct {
	serviceContext *services.ServiceContext
}

func NewExecCmdTestService(serviceContext *services.ServiceContext) *ExecCmdTestService {
	return &ExecCmdTestService{serviceContext: serviceContext}
}

func (self ExecCmdTestService) GetServiceContext() *services.ServiceContext {
	return self.serviceContext
}

func (self ExecCmdTestService) IsAvailable() bool {
	return true
}

