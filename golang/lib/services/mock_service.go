/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

const (
	MockServicePort = 1000
)

// Mock service, for testing purposes only
type MockService struct {
	// For testing, the service will report as available on the Nth call to IsAvailable
	becomesAvailableOnCheck int

	// Number of calls to IsAvailable that have happened
	callsToIsAvailable int
}

func NewMockService(becomesAvailableOnCheck int) *MockService {
	return &MockService{
		becomesAvailableOnCheck: becomesAvailableOnCheck,
		callsToIsAvailable:      0,
	}
}

func (m *MockService) IsAvailable() bool {
	m.callsToIsAvailable++
	return m.callsToIsAvailable >= m.becomesAvailableOnCheck
}

