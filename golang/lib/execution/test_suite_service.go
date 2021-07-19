package execution

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/docker_api/test_suite_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"os"
	"path"
	"sync"
)

type testSetupInfo struct {
	network networks.Network
	testName string
}

type TestSuiteService struct {
	// This embedding is required by gRPC
	bindings.UnimplementedTestSuiteServiceServer

	suite testsuite.TestSuite

	// This will only be non-empty after SetupTest is called
	testSetupInfo *testSetupInfo

	// Mutex to guard the testSetupInfo object, so any accidental concurrent calls of SetupInfo don't generate race conditions
	testSetupInfoMutex *sync.Mutex

	// Will only be non-nil if an IP:port to a Kurtosis API container was provided
	kurtosisApiClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient
}

func NewTestSuiteService(suite testsuite.TestSuite, kurtosisApiClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient) *TestSuiteService {
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
	staticFiles := service.suite.GetStaticFiles()
	staticFilesStrKeys := map[string]bool{}
	for key := range staticFiles {
		staticFilesStrKeys[string(key)] = true
	}
	testSuiteMetadata := &bindings.TestSuiteMetadata{
		TestMetadata:     allTestMetadata,
		NetworkWidthBits: networkWidthBits,
		StaticFiles: staticFilesStrKeys,
	}

	return testSuiteMetadata, nil
}

func (service *TestSuiteService) CopyStaticFilesToExecutionVolume(ctx context.Context, args *bindings.CopyStaticFilesToExecutionVolumeArgs) (*emptypb.Empty, error) {
	staticFileDestRelativeFilepaths := args.StaticFileDestRelativeFilepaths

	allStaticFiles := service.suite.GetStaticFiles()
	staticFileSrcAbsFilepaths := map[services.StaticFileID]string{}
	staticFileDestAbsFilepaths := map[services.StaticFileID]string{}
	for staticFileIdStr, destRelativeFilepath := range staticFileDestRelativeFilepaths {
		staticFileId := services.StaticFileID(staticFileIdStr)

		// Sanity-check that the source filepath exists
		srcAbsFilepath, found := allStaticFiles[staticFileId]
		if !found {
			return nil, stacktrace.NewError("The Kurtosis API gave a relative filepath for static file '%v', but the testsuite didn't declare a static file with this key!", staticFileId)
		}
		if _, err := os.Stat(srcAbsFilepath); os.IsNotExist(err) {
			return nil, stacktrace.NewError("Source filepath '%v' associated with static file '%v' doesn't exist", srcAbsFilepath, staticFileId)
		}
		staticFileSrcAbsFilepaths[staticFileId] = srcAbsFilepath

		// Sanity-check that a file has been created at the destination by Kurtosis
		destAbsFilepath := path.Join(test_suite_container_mountpoints.TestsuiteContainerSuiteExVolMountpoint, destRelativeFilepath)
		if _, err := os.Stat(destAbsFilepath); os.IsNotExist(err) {
			return nil, stacktrace.NewError("The Kurtosis API asked us to copy static file '%v' to path '%v' in the suite execution volume, but no file exists there - this is a bug in Kurtosis!", staticFileId, destRelativeFilepath)
		}
		staticFileDestAbsFilepaths[staticFileId] = destAbsFilepath
	}

	for staticFileIdStr := range staticFileDestRelativeFilepaths {
		staticFileId := services.StaticFileID(staticFileIdStr)

		srcAbsFilepath, found := staticFileSrcAbsFilepaths[staticFileId]
		if !found {
			return nil, stacktrace.NewError("No source filepath found for static file '%v'; this is a bug in Kurtosis", staticFileId)
		}
		destAbsFilepath, found := staticFileDestAbsFilepaths[staticFileId]
		if !found {
			return nil, stacktrace.NewError("No destination filepath found for static file '%v'; this is a bug in Kurtosis", staticFileId)
		}

		srcFp, err := os.Open(srcAbsFilepath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred opening static file '%v' source file '%v' for reading", staticFileId, srcAbsFilepath)
		}
		defer srcFp.Close()

		destFp, err := os.Create(destAbsFilepath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred opening static file '%v' destination file '%v' for writing", staticFileId, destAbsFilepath)
		}
		defer destFp.Close()

		if _, err := io.Copy(destFp, srcFp); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred copying all the bytes from static file '%v' source filepath '%v' to destination filepath '%v'", staticFileId, srcAbsFilepath, destAbsFilepath)
		}
	}

	return &emptypb.Empty{}, nil
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
		test_suite_container_mountpoints.TestsuiteContainerSuiteExVolMountpoint,
	)

	userNetwork, err := test.Setup(networkCtx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred during test setup")
	}
	if userNetwork == nil {
		return nil, stacktrace.NewError("The test setup method returned successfully, but yielded a nil network object - this is a bug with the test's setup method accidentally returning a nil network object")
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
		return nil, stacktrace.Propagate(
			err,
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
