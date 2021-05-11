package execution

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
)

type testSetupInfo struct {
	network networks.Network
	testName string
}

type TestSuiteService struct {
	suite testsuite.TestSuite

	// This will only be non-empty after SetupTest is called
	testSetupInfo *testSetupInfo

	testSetupInfoMutex *sync.Mutex

	// Will only be non-nil if an IP:port to a Kurtosis API container was provided
	kurtosisApiClient core_api_bindings.ApiContainerServiceClient
}

func NewTestSuiteService(suite testsuite.TestSuite, kurtosisApiClient core_api_bindings.ApiContainerServiceClient) *TestSuiteService {
	return &TestSuiteService{
		suite:              suite,
		testSetupInfo:      nil,
		testSetupInfoMutex: &sync.Mutex{},
		kurtosisApiClient:  kurtosisApiClient,
	}
}

func (service TestSuiteService) IsAvailable(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (service TestSuiteService) GetTestSuiteMetadata(ctx context.Context, empty *emptypb.Empty) (*bindings.TestSuiteMetadata, error) {
	allTestMetadata := map[string]*bindings.TestMetadata{}
	for testName, test := range service.suite.GetTests() {
		testConfigBuilder := testsuite.NewTestConfigurationBuilder()
		test.Configure(testConfigBuilder)
		testConfig := testConfigBuilder.Build()
		usedArtifactUrls := map[string]bool{}
		for _, artifactUrl := range testConfig.FilesArtifactUrls {
			usedArtifactUrls[artifactUrl] = true
		}
		testMetadata := &bindings.TestMetadata{
			IsPartitioningEnabled: testConfig.IsPartitioningEnabled,
			UsedArtifactUrls:      usedArtifactUrls,
			TestSetupTimeoutInSeconds: testConfig.SetupTimeoutSeconds,
			TestRunTimeoutInSeconds: testConfig.RunTimeoutSeconds,
		}
		allTestMetadata[testName] = testMetadata
	}

	networkWidthBits := service.suite.GetNetworkWidthBits()
	testSuiteMetadata := &bindings.TestSuiteMetadata{
		TestMetadata:     allTestMetadata,
		NetworkWidthBits: networkWidthBits,
	}

	return testSuiteMetadata, nil
}

func (service *TestSuiteService) SetupTest(ctx context.Context, args *bindings.SetupTestArgs) (*emptypb.Empty, error) {
	service.testSetupInfoMutex.Lock()
	defer service.testSetupInfoMutex.Unlock()

	if service.kurtosisApiClient == nil {
		return nil, stacktrace.NewError("Received a request to setup the test, but the Kurtosis API container client is nil")
	}

	testName := args.TestName

	allTests := service.suite.GetTests()
	test, found := allTests[testName]
	if !found {
		return nil, stacktrace.NewError(
			"Testsuite was directed to setup test '%v', but no test with that name exists " +
				"in the testsuite; this is a Kurtosis code bug",
			testName,
		)
	}

	logrus.Infof("Setting up network for test '%v'...", testName)
	testConfigBuilder := testsuite.NewTestConfigurationBuilder()
	test.Configure(testConfigBuilder)
	testConfig := testConfigBuilder.Build()
	filesArtifactUrls := testConfig.FilesArtifactUrls

	networkCtx := networks.NewNetworkContext(
		service.kurtosisApiClient,
		filesArtifactUrls,
	)

	userNetwork, err := test.Setup(networkCtx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred during test setup")
	}
	service.testSetupInfo = &testSetupInfo{
		network:  userNetwork,
		testName: testName,
	}
	logrus.Infof("Successfully set up test network for test '%v'", testName)

	return &emptypb.Empty{}, nil
}

func (service TestSuiteService) RunTest(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	service.testSetupInfoMutex.Lock()
	defer service.testSetupInfoMutex.Unlock()

	if service.kurtosisApiClient == nil {
		return nil, stacktrace.NewError("Received a request to run the test, but the Kurtosis API container client is nil")
	}
	if service.testSetupInfo == nil {
		return nil, stacktrace.NewError("Received a request to run the test, but the test hasn't been set up yet")
	}

	network := service.testSetupInfo.network
	testName := service.testSetupInfo.testName

	allTests := service.suite.GetTests()
	test, found := allTests[testName]
	if !found {
		return nil, stacktrace.NewError(
			"Testsuite was directed to run test '%v', but no test with that name exists " +
				"in the testsuite; this is a Kurtosis code bug",
			testName,
		)
	}

	logrus.Infof("Running test logic for test '%v'...", testName)
	if err := runTest(test, network); err != nil {
		return nil, stacktrace.NewError(
			"An error occurred running test '%v'",
			testName,
		)
	}
	logrus.Infof("Ran test logic for test '%v'", testName)
	return &emptypb.Empty{}, nil
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
