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
	serviceId ServiceID

	ipAddr string

	// For testing, the service will report as available on the Nth call to IsAvailable
	becomesAvailableOnCheck int

	// Number of calls to IsAvailable that have happened
	callsToIsAvailable int
}

func NewMockService(serviceId ServiceID, ipAddr string, becomesAvailableOnCheck int) *MockService {
	return &MockService{
		serviceId: serviceId,
		ipAddr:                  ipAddr,
		becomesAvailableOnCheck: becomesAvailableOnCheck,
		callsToIsAvailable:      0,
	}
}

func (m MockService) GetServiceID() ServiceID {
	return m.serviceId
}

func (m MockService) GetIPAddress() string {
	return m.ipAddr
}

func (m *MockService) IsAvailable() bool {
	m.callsToIsAvailable++
	return m.callsToIsAvailable >= m.becomesAvailableOnCheck
}

