package exec_command_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/exec_cmd_test"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	execCmdTestImage = "alpine:3.12.4"
	testServiceId = "test"

	successExitCode int32 = 0

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls = 10
)

var execCommandThatShouldWork = []string{
	"true",
}

var execCommandThatShouldFail = []string{
	"false",
}

type ExecCommandTest struct {}

func (e ExecCommandTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{}
}

func (e ExecCommandTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	initializer := exec_cmd_test.NewExecCmdTestContainerInitializer(execCmdTestImage)
	_, checker, err := networkCtx.AddService(testServiceId, initializer)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting service '%v'",
			testServiceId)
	}
	if err := checker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxPolls); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for service '%v' to start up",
			testServiceId)
	}
	return networkCtx, nil
}

func (e ExecCommandTest) Run(uncastedNetwork networks.Network, testCtx testsuite.TestContext) {
	network := uncastedNetwork.(*networks.NetworkContext)

	uncastedService, err := network.GetService(testServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting service with ID '%v'", testServiceId))
	}
	castedService := uncastedService.(*exec_cmd_test.ExecCmdTestService)

	logrus.Infof("Running exec command '%v' that should return a successful exit code...", execCommandThatShouldWork)
	shouldWorkExitCode, err := castedService.RunExecCmd(execCommandThatShouldWork)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldWork))
	}
	if shouldWorkExitCode != successExitCode {
		testCtx.Fatal(stacktrace.NewError("Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldWork, shouldWorkExitCode))
	}
	logrus.Info("Exec command returned successful exit code as expected")


	logrus.Infof("Running exec command '%v' that should return an error exit code...", execCommandThatShouldFail)
	shouldFailExitCode, err := castedService.RunExecCmd(execCommandThatShouldFail)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldFail))
	}
	if shouldFailExitCode == successExitCode {
		testCtx.Fatal(stacktrace.NewError("Exec command '%v' should fail, but got successful exit code %v", shouldFailExitCode, successExitCode))
	}

	// TODO TODO TODO Implement test for exec command that should return logs
	logrus.Info("Exec command returned error exit code as expected")
}

func (e ExecCommandTest) GetSetupTimeout() time.Duration {
	return 30 * time.Second
}

func (e ExecCommandTest) GetExecutionTimeout() time.Duration {
	return 30 * time.Second
}


