/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

type Person struct {
	BooksRead int
}

type ApiService struct {
	serviceCtx *services.ServiceContext
	port       int
}

func NewApiService(serviceCtx *services.ServiceContext, port int) *ApiService {
	return &ApiService{serviceCtx: serviceCtx, port: port}
}

func (service ApiService) GetServiceContext() *services.ServiceContext {
	return service.serviceCtx
}

func (service ApiService) IsAvailable() bool {
	return true
}

func (service ApiService) GetPort() int{
	return service.port
}