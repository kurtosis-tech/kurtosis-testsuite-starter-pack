/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package datastore

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

type DatastoreService struct {
	serviceCtx *services.ServiceContext
	port int
}

func NewDatastoreService(serviceCtx *services.ServiceContext, port int) *DatastoreService {
	return &DatastoreService{serviceCtx: serviceCtx, port: port}
}

func (service DatastoreService) GetServiceContext() *services.ServiceContext {
	return service.serviceCtx
}

func (service DatastoreService) IsAvailable() bool {
	return true
}

func (service DatastoreService) GetPort() int{
	return service.port
}
