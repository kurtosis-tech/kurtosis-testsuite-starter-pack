/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	maxSuiteRegistrationRetries = 20
	timeBetweenSuiteRegistrationRetries = 500 * time.Millisecond

	apiContainerConnTimeout = 30 * time.Second
)

type TestSuiteExecutor struct {
	kurtosisApiSocket string
	logLevelStr string
	paramsJsonStr string
	configurator TestSuiteConfigurator
}

func NewTestSuiteExecutor(kurtosisApiSocket string, logLevelStr string, paramsJsonStr string, configurator TestSuiteConfigurator) *TestSuiteExecutor {
	return &TestSuiteExecutor{kurtosisApiSocket: kurtosisApiSocket, logLevelStr: logLevelStr, paramsJsonStr: paramsJsonStr, configurator: configurator}
}

func (executor *TestSuiteExecutor) Run(ctx context.Context) error {
	if err := executor.configurator.SetLogLevel(executor.logLevelStr); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting the loglevel before running the testsuite")
	}

	suite, err := executor.configurator.ParseParamsAndCreateSuite(executor.paramsJsonStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the suite params JSON and creating the testsuite")
	}

	timeoutContext, cancelFunc := context.WithTimeout(context.Background(), apiContainerConnTimeout)
	defer cancelFunc()
	conn, err := grpc.DialContext(
		// Bit weird that the dial timeout is configured via a context, but that's what the docs instruct
		timeoutContext,
		executor.kurtosisApiSocket,
		grpc.WithInsecure(), // TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
		grpc.WithBlock(),	// This is required for the timeout context to take effect
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis API server at '%v'",
			executor.kurtosisApiSocket)
	}
	defer conn.Close()

	suiteRegistrationClient := core_api_bindings.NewSuiteRegistrationServiceClient(conn)

	var suiteRegistrationResp *core_api_bindings.SuiteRegistrationResponse
	suiteRegistrationAttempts := 0
	for {
		if suiteRegistrationAttempts >= maxSuiteRegistrationRetries {
			return stacktrace.NewError(
				"Failed to register testsuite with API container, even after %v retries spaced %v apart",
				maxSuiteRegistrationRetries,
				timeBetweenSuiteRegistrationRetries)
		}

		resp, err := suiteRegistrationClient.RegisterSuite(ctx, &emptypb.Empty{})
		if err == nil {
			suiteRegistrationResp = resp
			break
		}
		logrus.Debugf("The following error occurred registering testsuite with API container; retrying in %v: %v", timeBetweenSuiteRegistrationRetries, err)
		time.Sleep(timeBetweenSuiteRegistrationRetries)
		suiteRegistrationAttempts++
	}

	action := suiteRegistrationResp.SuiteAction
	switch action {
	case core_api_bindings.SuiteAction_SERIALIZE_SUITE_METADATA:
		if err := runSerializeSuiteMetadataFlow(ctx, suite, conn); err != nil {
			return stacktrace.Propagate(err, "An error occurred running the suite metadata serialization flow")
		}
		return nil
	case core_api_bindings.SuiteAction_EXECUTE_TEST:
		if err := runTestExecutionFlow(ctx, suite, conn); err != nil {
			return stacktrace.Propagate(err, "An error occurred running the test execution flow")
		}
		return nil
	default:
		return stacktrace.NewError("Encountered unrecognized action '%v'; this is a bug in Kurtosis itself", action)
	}
}

func runSerializeSuiteMetadataFlow(ctx context.Context, suite testsuite.TestSuite, conn *grpc.ClientConn) error {
	allTestMetadata := map[string]*core_api_bindings.TestMetadata{}
	for testName, test := range suite.GetTests() {
		testConfigBuilder := testsuite.NewTestConfigurationBuilder()
		test.Configure(testConfigBuilder)
		testConfig := testConfigBuilder.Build()
		usedArtifactUrls := map[string]bool{}
		for _, artifactUrl := range testConfig.FilesArtifactUrls {
			usedArtifactUrls[artifactUrl] = true
		}
		testMetadata := &core_api_bindings.TestMetadata{
			IsPartitioningEnabled: testConfig.IsPartitioningEnabled,
			UsedArtifactUrls:      usedArtifactUrls,
			TestSetupTimeoutInSeconds: testConfig.SetupTimeoutSeconds,
			TestRunTimeoutInSeconds: testConfig.RunTimeoutSeconds,
		}
		allTestMetadata[testName] = testMetadata
	}

	networkWidthBits := suite.GetNetworkWidthBits()
	testSuiteMetadata := &core_api_bindings.TestSuiteMetadata{
		TestMetadata:     allTestMetadata,
		NetworkWidthBits: networkWidthBits,
	}

	metadataSerializationClient := core_api_bindings.NewSuiteMetadataSerializationServiceClient(conn)
	if _, err := metadataSerializationClient.SerializeSuiteMetadata(ctx, testSuiteMetadata); err != nil {
		return stacktrace.Propagate(err, "An error occurred sending the suite metadata to the Kurtosis API server")
	}

	return nil
}

func runTestExecutionFlow(ctx context.Context, suite testsuite.TestSuite, conn *grpc.ClientConn) error {
	executionClient := core_api_bindings.NewTestExecutionServiceClient(conn)
	testExecutionInfo, err := executionClient.GetTestExecutionInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test execution info")
	}
	testName := testExecutionInfo.TestName

	allTests := suite.GetTests()
	test, found := allTests[testName]
	if !found {
		return stacktrace.NewError(
			"Testsuite was directed to execute test '%v', but no test with that name exists " +
				"in the testsuite; this is a Kurtosis code bug",
			testName)
	}

	testConfigBuilder := testsuite.NewTestConfigurationBuilder()
	test.Configure(testConfigBuilder)
	testConfig := testConfigBuilder.Build()
	filesArtifactUrls := testConfig.FilesArtifactUrls

	networkCtx := networks.NewNetworkContext(
		executionClient,
		filesArtifactUrls)

	logrus.Info("Setting up the test network...")
	// Kick off a timer with the API in case there's an infinite loop in the user code that causes the test to hang forever
	logrus.Debug("Registering test setup with API container...")
	if _, err := executionClient.RegisterTestSetup(ctx, &emptypb.Empty{}); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the test setup with the API container")
	}
	logrus.Debug("Test setup registered with API container")
	logrus.Debug("Executing setup logic...")
	untypedNetwork, err := test.Setup(networkCtx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test network")
	}
	logrus.Debug("Setup logic executed")
	logrus.Debug("Registering test setup completion...")
	if _, err := executionClient.RegisterTestSetupCompletion(ctx, &emptypb.Empty{}); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the test setup completion with the API container")
	}
	logrus.Debug("Test setup completion registered")
	logrus.Info("Test network set up")

	logrus.Infof("Executing test '%v'...", testName)
	if _, err := executionClient.RegisterTestExecution(ctx, &emptypb.Empty{}); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the test execution with the API container")
	}
	testResultErr := runTest(test, untypedNetwork)
	logrus.Tracef("After running test: resultErr: %v", testResultErr)
	logrus.Infof("Executed test '%v'", testName)

	if testResultErr != nil {
		return stacktrace.Propagate(testResultErr, "An error occurred when running the test")
	}

	return nil
}

// Little helper function that runs the test and captures panics on test failures, returning them as errors
func runTest(test testsuite.Test, untypedNetwork interface{}) (resultErr error) {
	// See https://medium.com/@hussachai/error-handling-in-go-a-quick-opinionated-guide-9199dd7c7f76 for details
	defer func() {
		if recoverResult := recover(); recoverResult != nil {
			logrus.Tracef("Caught panic while running test: %v", recoverResult)
			resultErr = recoverResult.(error)
		}
	}()
	if err := test.Run(untypedNetwork); err != nil {
		return stacktrace.Propagate(err, "The test returned an error")
	}
	logrus.Tracef("Test completed successfully")
	return
}
