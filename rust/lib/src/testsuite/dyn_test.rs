use std::{borrow::Borrow, ops::Deref};

use crate::{core_api_bindings::api_container_api::{RegisterTestExecutionArgs, suite_registration_service_client, test_execution_service_client::TestExecutionServiceClient, test_execution_service_server::TestExecutionService}, networks::network_context::NetworkContext};

use super::{test::Test, test_configuration::TestConfiguration, test_context::TestContext};
use anyhow::{Context, Result};
use futures::executor::block_on;
use log::info;
use tonic::transport::Channel;

// This trait is necessary to genericize across tests, so that they can be put inside a HashMap
// since Rust doesn't allow things like HashMap<String, Test<? extends Network>>
// See: https://discord.com/channels/442252698964721669/448238009733742612/809977090740977674
pub trait DynTest {
    fn get_test_configuration(&self) -> TestConfiguration;

    fn setup_and_run(&self, channel: Channel) -> Result<()>;
}

// This struct exists to shield the genericized N parameter
// See: https://discord.com/channels/442252698964721669/448238009733742612/809977090740977674
pub struct DynTestContainer<T: Test> {
    test: T,
}

impl<T: Test> DynTestContainer<T> {
    pub fn new(test: T) -> DynTestContainer<T> {
        return DynTestContainer{
            test,
        };
    }


/*
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

 */
}

impl<T: Test> DynTest for DynTestContainer<T> {
    fn get_test_configuration(&self) -> TestConfiguration {
        return self.test.get_test_configuration();
    }
    
    fn setup_and_run(&self, channel: Channel) -> Result<()> {
        // TODO create NetworkContext????
        let test_config = self.test.get_test_configuration();
        let files_artifact_urls = test_config.files_artifact_urls;
        // It's weird that we're cloning the channel, but this is how you're supposed to do it according to the
        // Channel documentation since it uses a &mut self
        let network_ctx_client = TestExecutionServiceClient::new(channel.clone());
        let network_ctx = NetworkContext::new(network_ctx_client, files_artifact_urls);
        let mut registration_client = TestExecutionServiceClient::new(channel.clone());

		// TODO this needs to be refactored to the new world!!!
        let test_execution_timeout = self.test.get_execution_timeout() + self.test.get_setup_teardown_buffer();
        let register_test_execution_req = tonic::Request::new(RegisterTestExecutionArgs{
            timeout_seconds: test_execution_timeout.as_secs(),
        });
        block_on(registration_client.register_test_execution(register_test_execution_req))
            .context("An error occurred registering the test execution with the API container")?;

        info!("Setting up the test network...");
        // TODO register setup
        let network = self.test.setup(network_ctx)
            .context("An error occurred setting up the test network")?;
        // TODO register setup completion
        info!("Test network set up");

        let test_ctx = TestContext{};

        info!("Executing the test...");

        self.test.run(network, test_ctx)
            .context("An error occurred executing the test")?;
        info!("Test execution completed");

        return Ok(());
	/*

		// Kick off a timer with the API in case there's an infinite loop in the user code that causes the test to hang forever
		// TODO this should just be "register test execution started", since the API container already has the metadata
		hardTestTimeout := test.GetExecutionTimeout() + test.GetSetupTeardownBuffer()
		hardTestTimeoutSeconds := uint64(hardTestTimeout.Seconds())
		registerTestExecutionMessage := &core_api_bindings.RegisterTestExecutionArgs{TimeoutSeconds: hardTestTimeoutSeconds}
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
	*/
    }
}