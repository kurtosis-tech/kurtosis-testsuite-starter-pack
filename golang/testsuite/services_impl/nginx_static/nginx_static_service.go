/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package nginx_static

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
)

/*
An Nginx service that serves files mounted in the /static directory
 */
type NginxStaticService struct {
	serviceCtx *services.ServiceContext
}

func NewNginxStaticService(serviceCtx *services.ServiceContext) *NginxStaticService {
	return &NginxStaticService{serviceCtx: serviceCtx}
}

func (self NginxStaticService) GetServiceContext() *services.ServiceContext {
	return self.serviceCtx
}

func (self NginxStaticService) IsAvailable() bool {
	return true
}



