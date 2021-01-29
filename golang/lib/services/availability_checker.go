/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

import (
	"github.com/palantir/stacktrace"
	"time"
)

type AvailabilityChecker interface {
	WaitForStartup(timeBetweenPolls time.Duration, maxNumRetries int) error
}

/*
Struct for polling a service until it's available, with configurable retry options
 */
type DefaultAvailabilityChecker struct {
	// ID of the service being monitored
	serviceId ServiceID

	// The service being monitored
	toCheck Service
}

func NewDefaultAvailabilityChecker(serviceId ServiceID, toCheck Service) *DefaultAvailabilityChecker {
	return &DefaultAvailabilityChecker{serviceId: serviceId, toCheck: toCheck}
}

/*
Waits for the service that was passed in at construction time to start up by making requests to the service until
	the service is available or the maximum number of retries are reached
 */
func (checker DefaultAvailabilityChecker) WaitForStartup(timeBetweenPolls time.Duration, maxNumRetries int) error {
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
}
