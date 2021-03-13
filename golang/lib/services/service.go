/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package services

/*
The identifier used for services with the network.
*/
type ServiceID string

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
type Service interface {
	// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
	IsAvailable() bool
}
