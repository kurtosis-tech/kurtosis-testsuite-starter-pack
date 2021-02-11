use std::collections::HashMap;

use kurtosis_rust_lib::testsuite::testsuite::TestSuite;

pub struct ExampleTestsuite {}

impl ExampleTestsuite {
    pub fn new() -> ExampleTestsuite {
        return ExampleTestsuite{};
    }
}

impl TestSuite for ExampleTestsuite {
    fn get_tests(&self) -> std::collections::HashMap<String, Box<dyn kurtosis_rust_lib::testsuite::test::Test<dyn kurtosis_rust_lib::networks::network::Network>>> {
        // TODO replace with actual tests
        return HashMap::new();
    }

    fn get_network_width_bits(&self) -> u32 {
        return 8;
    }
}