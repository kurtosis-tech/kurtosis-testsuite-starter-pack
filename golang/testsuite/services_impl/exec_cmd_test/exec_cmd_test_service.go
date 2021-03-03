package exec_cmd_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
)

type ExecCmdTestService struct {
	serviceContext *services.ServiceContext
}

func NewExecCmdTestService(serviceContext *services.ServiceContext) *ExecCmdTestService {
	return &ExecCmdTestService{serviceContext: serviceContext}
}

func (self ExecCmdTestService) IsAvailable() bool {
	return true;
}

func (self ExecCmdTestService) RunExecCmd(command []string) (int32, error) {
	exitCode, err := self.serviceContext.ExecCommand(command)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred executing command '%v'", command)
	}
	return exitCode, nil
}
