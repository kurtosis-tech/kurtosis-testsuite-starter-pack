use std::{rc::Rc, time::Duration};

use anyhow::{anyhow, Result};
use kurtosis_rust_lib::services::{availability_checker::AvailabilityChecker, service::{Service, ServiceId}};

struct MockService {
    is_available: bool,
}

impl MockService {
    pub fn new(is_available: bool) -> MockService {
        return MockService{
            is_available,
        };
    }
}

impl Service for MockService {
    fn is_available(&self) -> bool {
        return self.is_available;
    }
}

#[test]
fn test_timeout_on_service_startup() -> Result<()> {
    let never_available_service = MockService::new(false);
    let service_id: ServiceId = String::from("test-service");
    let checker = AvailabilityChecker::new(&service_id, Rc::new(never_available_service));

    match checker.wait_for_startup(&Duration::from_millis(200), 3) {
        Err(_) => Ok(()),
        Ok(_) => Err(anyhow!(
            "Expected an error waiting for a never-available service, but no error was thrown"
        )),
    }
}