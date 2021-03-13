use std::{rc::Rc, thread::sleep, time::Duration};
use anyhow::{anyhow, Result};

use super::service::Service;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct AvailabilityChecker {
    service_id: String,

    to_check: Rc<dyn Service>,
}

impl AvailabilityChecker {
    pub fn new(service_id: &str, to_check: Rc<dyn Service>) -> AvailabilityChecker {
        return AvailabilityChecker{
            service_id: service_id.to_owned(),
            to_check,
        };
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn wait_for_startup(&self, time_between_polls: &Duration, max_num_retries: u32) -> Result<()> {
        for i in 0..max_num_retries {
            if self.to_check.is_available() {
                return Ok(());
            }

            // Don't wait if we're on the last iteration of the loop, since we'd be waiting unnecessarily
            if i < max_num_retries - 1 {
                sleep(time_between_polls.to_owned());
            }
        }
        return Err(anyhow!(
            "Service '{}' did not become available despite polling {} times with {}ms between polls",
            self.service_id,
            max_num_retries,
            time_between_polls.as_millis(),
        ));
    }
}