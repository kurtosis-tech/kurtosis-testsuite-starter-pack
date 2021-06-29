package exec_command_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/exec_cmd_test"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	execCmdTestImage      = "alpine:3.12.4"
	inputForLogOutputTest = "hello"
	expectedLogOutput     = "hello\n"
	testServiceId         = "test"

	successExitCode int32 = 0

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls         = 10
)

var execCommandThatShouldWork = []string{
	"true",
}

var execCommandThatShouldHaveLogOutput = []string{
	"echo",
	inputForLogOutputTest,
}

var execCommandThatShouldFail = []string{
	"false",
}

type ExecCommandTest struct{}

func (e ExecCommandTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(30).WithRunTimeoutSeconds(30)
}

func (e ExecCommandTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	configFactory := exec_cmd_test.NewExecCmdTestContainerConfigFactory(execCmdTestImage)
	_, _, checker, err := networkCtx.AddService(testServiceId, configFactory)
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

func (e ExecCommandTest) Run(uncastedNetwork networks.Network) error {
	network := uncastedNetwork.(*networks.NetworkContext)

	uncastedService, err := network.GetService(testServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service with ID '%v'", testServiceId)
	}
	castedService := uncastedService.(*exec_cmd_test.ExecCmdTestService)

	logrus.Infof("Running exec command '%v' that should return a successful exit code...", execCommandThatShouldWork)
	shouldWorkExitCode, _, err := castedService.RunExecCmd(execCommandThatShouldWork)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldWork)
	}
	if shouldWorkExitCode != successExitCode {
		return stacktrace.NewError("Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldWork, shouldWorkExitCode)
	}
	logrus.Info("Exec command returned successful exit code as expected")

	logrus.Infof("Running exec command '%v' that should return an error exit code...", execCommandThatShouldFail)
	shouldFailExitCode, _, err := castedService.RunExecCmd(execCommandThatShouldFail)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldFail)
	}
	if shouldFailExitCode == successExitCode {
		return stacktrace.NewError("Exec command '%v' should fail, but got successful exit code %v", execCommandThatShouldFail, successExitCode)
	}

	logrus.Infof("Running exec command '%v' that should return log output...", execCommandThatShouldHaveLogOutput)
	shouldHaveLogOutputExitCode, logOutput, err := castedService.RunExecCmd(execCommandThatShouldHaveLogOutput)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldHaveLogOutput)
	}
	if shouldHaveLogOutputExitCode != successExitCode {
		return stacktrace.NewError("Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldHaveLogOutput, shouldHaveLogOutputExitCode)
	}
	logOutputStr := fmt.Sprintf("%s", *logOutput)
	if logOutputStr != expectedLogOutput {
		return stacktrace.NewError("Exec command '%v' should return %v, but got %v.", execCommandThatShouldHaveLogOutput, inputForLogOutputTest, logOutputStr)
	}
	logrus.Info("Exec command returned error exit code as expected")

	return nil
}
