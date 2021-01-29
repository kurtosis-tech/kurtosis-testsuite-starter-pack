/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api/bindings"
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

	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	conn, err := grpc.Dial(executor.kurtosisApiSocket, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis API server at '%v'",
			executor.kurtosisApiSocket)
	}
	defer conn.Close()

	suiteRegistrationClient := bindings.NewSuiteRegistrationServiceClient(conn)

	var suiteRegistrationResp *bindings.SuiteRegistrationResponse
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
		logrus.Debugf("The following error occurred registering testsuite with API container; retrying in %v:", timeBetweenSuiteRegistrationRetries)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		time.Sleep(timeBetweenSuiteRegistrationRetries)
		suiteRegistrationAttempts++
	}

	action := suiteRegistrationResp.SuiteAction
	switch action {
	case bindings.SuiteAction_SERIALIZE_SUITE_METADATA:
		if err := runSerializeSuiteMetadataFlow(ctx, suite, conn); err != nil {
			return stacktrace.Propagate(err, "An error occurred running the suite metadata serialization flow")
		}
		return nil
	case bindings.SuiteAction_EXECUTE_TEST:
		if err := runTestExecutionFlow(ctx, suite, conn); err != nil {
			return stacktrace.Propagate(err, "An error occurred running the test execution flow")
		}
		return nil
	default:
		return stacktrace.NewError("Encountered unrecognized action '%v'; this is a bug in Kurtosis itself", action)
	}
}

func runSerializeSuiteMetadataFlow(ctx context.Context, testsuite testsuite.TestSuite, conn *grpc.ClientConn) error {
	allTestMetadata := map[string]*bindings.TestMetadata{}
	for testName, test := range testsuite.GetTests() {
		testConfig := test.GetTestConfiguration()
		usedArtifactUrls := map[string]bool{}
		for _, artifactUrl := range testConfig.FilesArtifactUrls {
			usedArtifactUrls[artifactUrl] = true
		}

		testMetadata := &bindings.TestMetadata{
			IsPartitioningEnabled: testConfig.IsPartitioningEnabled,
			UsedArtifactUrls:      usedArtifactUrls,
		}
		allTestMetadata[testName] = testMetadata
	}

	networkWidthBits := testsuite.GetNetworkWidthBits()
	testSuiteMetadata := &bindings.TestSuiteMetadata{
		TestMetadata:     allTestMetadata,
		NetworkWidthBits: networkWidthBits,
	}

	metadataSerializationClient := bindings.NewSuiteMetadataSerializationServiceClient(conn)
	if _, err := metadataSerializationClient.SerializeSuiteMetadata(ctx, testSuiteMetadata); err != nil {
		return stacktrace.Propagate(err, "An error occurred sending the suite metadata to the Kurtosis API server")
	}

	return nil
}

func runTestExecutionFlow(ctx context.Context, testsuite testsuite.TestSuite, conn *grpc.ClientConn) error {
	executionClient := bindings.NewTestExecutionServiceClient(conn)
	testExecutionInfo, err := executionClient.GetTestExecutionInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test execution info")
	}
	testName := testExecutionInfo.TestName

	allTests := testsuite.GetTests()
	test, found := allTests[testName]
	if !found {
		return stacktrace.NewError(
			"Testsuite was directed to execute test '%v', but no test with that name exists " +
				"in the testsuite; this is a Kurtosis code bug",
			testName)
	}

	// Kick off a timer with the API in case there's an infinite loop in the user code that causes the test to hang forever
	// TODO this should just be "register test execution started", since the API container already has the metadata
	hardTestTimeout := test.GetExecutionTimeout() + test.GetSetupTeardownBuffer()
	hardTestTimeoutSeconds := uint64(hardTestTimeout.Seconds())
	registerTestExecutionMessage := &bindings.RegisterTestExecutionArgs{TimeoutSeconds: hardTestTimeoutSeconds}
	if _, err := executionClient.RegisterTestExecution(ctx, registerTestExecutionMessage); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the test execution with the API container")
	}

	testConfig := test.GetTestConfiguration()
	filesArtifactUrls := testConfig.FilesArtifactUrls

	networkCtx := networks.NewNetworkContext(
		executionClient,
		filesArtifactUrls)

	// TODO Also time out the setup with the API container rather than storing this locally
	//  to reduce complexity inside the lib
	logrus.Info("Setting up the test network...")
	untypedNetwork, err := test.Setup(networkCtx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test network")
	}
	logrus.Info("Test network set up")

	logrus.Infof("Executing test '%v'...", testName)
	testResultChan := make(chan error)

	go func() {
		testResultChan <- runTestInGoroutine(test, untypedNetwork)
	}()

	// TODO Switch to registering the timeout with the API container rather than storing this locally
	//  to reduce complexity inside the lib
	// Time out the test so a poorly-written test doesn't run forever
	testTimeout := test.GetExecutionTimeout()
	var timedOut bool
	var testResultErr error
	select {
	case testResultErr = <- testResultChan:
		logrus.Tracef("Test returned result before timeout: %v", testResultErr)
		timedOut = false
	case <- time.After(testTimeout):
		logrus.Tracef("Hit timeout %v before getting a result from the test", testTimeout)
		timedOut = true
	}
	logrus.Tracef("After running test w/timeout: resultErr: %v, timedOut: %v", testResultErr, timedOut)

	if timedOut {
		return stacktrace.NewError("Timed out after %v waiting for test to complete", testTimeout)
	}
	logrus.Infof("Executed test '%v'", testName)

	if testResultErr != nil {
		return stacktrace.Propagate(testResultErr, "An error occurred when running the test")
	}

	return nil
}

// Little helper function meant to be run inside a goroutine that runs the test
func runTestInGoroutine(test testsuite.Test, untypedNetwork interface{}) (resultErr error) {
	// See https://medium.com/@hussachai/error-handling-in-go-a-quick-opinionated-guide-9199dd7c7f76 for details
	defer func() {
		if recoverResult := recover(); recoverResult != nil {
			logrus.Tracef("Caught panic while running test: %v", recoverResult)
			resultErr = recoverResult.(error)
		}
	}()
	test.Run(untypedNetwork, testsuite.TestContext{})
	logrus.Tracef("Test completed successfully")
	return
}
