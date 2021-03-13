/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

import (
	"github.com/palantir/stacktrace"
	"time"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type AvailabilityChecker interface {
	WaitForStartup(timeBetweenPolls time.Duration, maxNumRetries int) error
}

type DefaultAvailabilityChecker struct {
	// ID of the service being monitored
	serviceId ServiceID

	// The service being monitored
	toCheck Service
}

func NewDefaultAvailabilityChecker(serviceId ServiceID, toCheck Service) *DefaultAvailabilityChecker {
	return &DefaultAvailabilityChecker{serviceId: serviceId, toCheck: toCheck}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
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
