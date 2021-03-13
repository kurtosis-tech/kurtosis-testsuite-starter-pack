use std::{collections::HashMap};

use super::{dyn_test::{DynTest}};

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub trait TestSuite {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	// NOTE: Due to Rust peculiarities, we need to return a result with instances of the `DynTest` object
	// rather than `Test`
	fn get_tests(&self) -> HashMap<String, Box<dyn DynTest>>;

	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	fn get_network_width_bits(&self) -> u32;
}