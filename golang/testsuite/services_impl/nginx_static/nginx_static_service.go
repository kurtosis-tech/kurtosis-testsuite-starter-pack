/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package nginx_static

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net/http"
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

func (self NginxStaticService) IsAvailable() bool {
	_, err := http.Get(fmt.Sprintf("%v:%v", self.serviceCtx.GetIPAddress(), listenPort))
	return err != nil
}

func (self NginxStaticService) GetFileContents(filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%v:%v/%v", self.serviceCtx.GetIPAddress(), listenPort, filename))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", filename)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body);
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}

