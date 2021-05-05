use async_trait::async_trait;
use std::{rc::Rc, sync::{Arc, Mutex}};

use tonic::transport::Channel;

use crate::{core_api_bindings::api_container_api::api_container_service_client::ApiContainerServiceClient, networks::network::Network, testsuite::testsuite::TestSuite};

struct TestSetupInfo {
    network: Rc<dyn Network>,
    test_name: String,
}

struct TestSuiteService {
    suite: Arc<dyn TestSuite>,

    // This will only be non-empty after setup is called
    test_setup_info: Arc<Mutex<Option<TestSetupInfo>>>,

	// Will only be non-nil if an IP:port to a Kurtosis API container was provided
    kurtosis_api_client: Option<ApiContainerServiceClient<Channel>>,
}

impl TestSuiteService {
    pub fn new(suite: Box<dyn TestSuite>, kurtosis_api_client: Option<ApiContainerServiceClient<Channel>>) -> TestSuiteService {
        return TestSuiteService{
            suite,
            test_setup_info: Arc::new(Mutex::new(None)),
            kurtosis_api_client,
        };
    }
}

#[async_trait]
impl crate::rpc_api::bindings::test_suite_api::test_suite_service_server::TestSuiteService for TestSuiteService {
    async fn get_test_suite_metadata(
            &self,
            request: tonic::Request<()>,
        ) -> Result<tonic::Response<crate::rpc_api::bindings::test_suite_api::TestSuiteMetadata>, tonic::Status> {
        
            /*
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

             */
    }

    async fn setup_test(
            &self,
            request: tonic::Request<crate::rpc_api::bindings::test_suite_api::SetupTestArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
        todo!()
    }

    async fn run_test(
            &self,
            request: tonic::Request<()>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
        todo!()
    }
}