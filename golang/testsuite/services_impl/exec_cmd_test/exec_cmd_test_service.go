package exec_cmd_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

func (self ExecCmdTestService) RunExecCmd(command []string) (int32, *[]byte, error) {
	exitCode, logOutput, err := self.serviceContext.ExecCommand(command)
	logrus.Infof("Exec Cmd Log Output in exec_cmd_test_service: %+v", logOutput)
	if err != nil {
		return 0, nil, stacktrace.Propagate(
			err,
			"An error occurred executing command '%v'", command)
	}
	return exitCode, logOutput, nil
}
