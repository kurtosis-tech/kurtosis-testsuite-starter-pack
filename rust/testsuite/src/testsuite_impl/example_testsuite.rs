use std::{any::Any, collections::HashMap};

use kurtosis_rust_lib::{networks::network::Network, testsuite::{dyn_test::{DynTest, DynTestContainer}, test::Test, testsuite::TestSuite}};

use super::basic_datastore_test::BasicDatastoreTest;

pub struct ExampleTestsuite {
    datastore_service_image: String,
}

impl ExampleTestsuite {
    pub fn new(datastore_service_image: String) -> ExampleTestsuite {
        return ExampleTestsuite{
            datastore_service_image,
        };
    }
}

impl TestSuite for ExampleTestsuite {
    fn get_tests(&self) -> HashMap<String, Box<dyn DynTest>> {
        let mut result: HashMap<String, Box<dyn DynTest>> = HashMap::new();

        let basic_datastore_test = BasicDatastoreTest::new(&self.datastore_service_image);
        let basic_datastore_test_container = DynTestContainer::new(basic_datastore_test);

        result.insert(
            "basicDatastoreTest".to_owned(), 
            Box::new(basic_datastore_test_container),
        );
        return result;
    }

    fn get_network_width_bits(&self) -> u32 {
        return 8;
    }
}