package local_static_file_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/exec_cmd_test"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/testsuite_impl/static_file_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	dockerImage                    = "alpine:3.12.4"
	testService services.ServiceID = "test-service"

	execCommandSuccessExitCode = 0
	expectedTestFileContents = "This is a test static file"
)

type LocalStaticFileTest struct {}

func (l LocalStaticFileTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (l LocalStaticFileTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	configFactory := exec_cmd_test.NewExecCmdTestContainerConfigFactory(dockerImage)
	_, _, _, err := networkCtx.AddService(testService, configFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}
	return networkCtx, nil
}

func (l LocalStaticFileTest) Run(network networks.Network) error {
	castedNetwork, ok := network.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred casting the uncasted network to a NetworkContext")
	}

	uncastedService, err := castedNetwork.GetService(testService)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service '%v'", testService)
	}
	castedService, ok := uncastedService.(*exec_cmd_test.ExecCmdTestService)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the uncasted service")
	}

	serviceCtx := castedService.GetServiceContext()
	staticFileAbsFilepaths, err := serviceCtx.LoadStaticFiles(map[services.StaticFileID]bool{static_file_consts.TestStaticFileID: true})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the static file corresponding to key '%v'", static_file_consts.TestStaticFileID)
	}
	testFileAbsFilepath, found := staticFileAbsFilepaths[static_file_consts.TestStaticFileID]
	if !found {
		return stacktrace.Propagate(err, "No filepath found for test file key '%v'; this is a bug in Kurtosis!", static_file_consts.TestStaticFileID)
	}

	catStaticFileCmd := []string{
		"cat",
		testFileAbsFilepath,
	}
	exitCode, outputBytes, err := serviceCtx.ExecCommand(catStaticFileCmd)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static test file contents", catStaticFileCmd)
	}
	if exitCode != execCommandSuccessExitCode {
		return stacktrace.NewError("Command '%+v' to cat the static test file exited with non-successful exit code '%v'", catStaticFileCmd, exitCode)
	}
	fileContents := string(*outputBytes)

	if fileContents != expectedTestFileContents {
		return stacktrace.NewError("Static file contents '%v' don't match expected test file contents '%v'", fileContents, expectedTestFileContents)
	}
	logrus.Infof("Static file contents were '%v' as expected", expectedTestFileContents)
	return nil
}