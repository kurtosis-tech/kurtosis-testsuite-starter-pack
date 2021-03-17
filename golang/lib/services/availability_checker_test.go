/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

import (
	"testing"
	"time"
)

const (
	testServiceId ServiceID = "test-service"
)

func TestServiceBecomesAvailable(t *testing.T) {
	service := NewMockService(2)
	availabilityChecker := NewDefaultAvailabilityChecker(testServiceId, service)

	if err := availabilityChecker.WaitForStartup(200 * time.Millisecond, 3); err != nil {
		t.Fatalf("Expected service to become available successfully but an error was thrown: %v", err)
	}
}

func TestTimeoutOnServiceStartup(t *testing.T) {
	neverAvailableService := NewMockService(9999)
	availabilityChecker := NewDefaultAvailabilityChecker(testServiceId, neverAvailableService)

	if err := availabilityChecker.WaitForStartup(200 * time.Millisecond, 3); err == nil {
		t.Fatalf("Expected an error waiting for a never-available service, but no error was thrown")
	}
}
