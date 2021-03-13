use core::panic;

use anyhow::Error;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct TestContext {}

impl TestContext {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn fatal(&self, err: Error) {
        TestContext::fail_test(err);
    }

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn assert_true(&self, condition: bool, err: Error) {
        if !condition {
            TestContext::fail_test(err);
        }
    }

    fn fail_test(err: Error) {
        panic!("{:?}", err);
    }
}