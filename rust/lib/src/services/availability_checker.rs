use std::time::Duration;

pub struct AvailabilityChecker {
    service_id: &str,

    to_check: &dyn Service,
}

impl AvailabilityChecker {
    pub fn wait_for_startup(time_between_polls: &Duration, max_num_retries: u32) -> Result<()> {
        // TODO
        /*
        	for i := 0; i < maxNumRetries; i++ {
		if checker.toCheck.IsAvailable() {
			return nil
		}

		// Don't wait if we're on the last iteration of the loop, since we'd be waiting unnecessarily
		if i < maxNumRetries - 1 {
			time.Sleep(timeBetweenPolls)
		}
	}
	return stacktrace.NewError(
		"Service '%v' did not become available despite polling %v times with %v between polls",
		checker.serviceId,
		maxNumRetries,
		timeBetweenPolls)
 */
        Ok(());
    }
}