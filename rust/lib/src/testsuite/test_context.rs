use core::panic;

use anyhow::Error;


/*
An object that will be passed in to every test, which the user can use to manipulate the results of the test
 */
pub struct TestContext {}

impl TestContext {
    /*
    Fails the test with the given error
    */
    pub fn fatal(&self, err: Error) {
        TestContext::fail_test(err);
    }

    /*
    Asserts that the given condition is true, and if not then fails the test and returns the given error
    */
    pub fn assert_true(&self, condition: bool, err: Error) {
        if !condition {
            TestContext::fail_test(err);
        }
    }

    fn fail_test(err: Error) {
        panic!("{:?}", err);
    }
}