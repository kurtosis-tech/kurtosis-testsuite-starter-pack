/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package kurtosis_service

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/palantir/stacktrace"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	kurtosisApiPort = 7443

	registrationRetryDurationSeconds = 60
	regularOperationRetryDurationSeconds = 10

	kurtosisServiceStruct = "KurtosisService"
	addServiceMethod = kurtosisServiceStruct + ".AddService"
	removeServiceMethod = kurtosisServiceStruct + ".RemoveService"
	registerTestExecutionMethod = kurtosisServiceStruct + ".RegisterTestExecution"
)

type KurtosisService struct {
	ipAddr string
}

func NewKurtosisService(ipAddr string) *KurtosisService {
	return &KurtosisService{ipAddr: ipAddr}
}

/*
Calls the Kurtosis API container to add a service to the network
 */
func (service KurtosisService) AddService(
		dockerImage string,
		// TODO change type of this to be an actual Port type
		usedPorts map[int]bool,
		ipPlaceholder string,
		startCmdArgs []string,
		envVariables map[string]string,
		testVolumeMountLocation string) (string, string, error) {
	client := getConstantBackoffJsonRpcClient(service.ipAddr, regularOperationRetryDurationSeconds)
	defer client.Close()

	usedPortsList := []int{}
	for port, _ := range usedPorts {
		usedPortsList = append(usedPortsList, port)
	}
	args := AddServiceArgs{
		IPPlaceholder: ipPlaceholder,
		ImageName:               dockerImage,
		UsedPorts:               usedPortsList,
		StartCmd:                startCmdArgs,
		DockerEnvironmentVars:   envVariables,
		TestVolumeMountFilepath: testVolumeMountLocation,
	}
	var reply AddServiceResponse
	if err := client.Call(addServiceMethod, args, &reply); err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred making the call to add a service using the Kurtosis API")
	}

	return reply.IPAddress, reply.ContainerID, nil
}

/*
Stops the container with the given service ID, and removes it from the network.
*/
func (service KurtosisService) RemoveService(containerId string, containerStopTimeoutSeconds int) error {
	client := getConstantBackoffJsonRpcClient(service.ipAddr, regularOperationRetryDurationSeconds)
	defer client.Close()

	logrus.Debugf("Removing service with container ID %v...", containerId)

	args := RemoveServiceArgs{
		ContainerID: containerId,
		ContainerStopTimeoutSeconds: containerStopTimeoutSeconds,
	}

	var reply struct{}
	if err := client.Call(removeServiceMethod, args, &reply); err != nil {
		return stacktrace.Propagate(err, "An error occurred making the call to remove a service using the Kurtosis API")
	}
	logrus.Debugf("Successfully removed service with container ID %v", containerId)

	return nil
}

func (service KurtosisService) RegisterTestExecution(testTimeoutSeconds int) error {
	client := getConstantBackoffJsonRpcClient(service.ipAddr, registrationRetryDurationSeconds)
	defer client.Close()

	logrus.Debugf("Registering a test execution with a timeout of %v seconds...", testTimeoutSeconds)

	args := RegisterTestExecutionArgs{TestTimeoutSeconds: testTimeoutSeconds}

	var reply struct{}
	if err := client.Call(registerTestExecutionMethod, args, &reply); err != nil {
		return stacktrace.Propagate(err, "An error occurred making the call to register a test execution using the Kurtosis API")
	}
	logrus.Debugf("Successfully registered a test execution with timeout of %v seconds", testTimeoutSeconds)

	return nil

}

// ================================= Private helper function ============================================
func getConstantBackoffJsonRpcClient(ipAddr string, retryDurationSeconds int) *jsonrpc2.Client {
	kurtosisUrl := fmt.Sprintf("http://%v:%v", ipAddr, kurtosisApiPort)
	retryingClient := retryablehttp.NewClient()
	retryingClient.RetryMax = retryDurationSeconds
	retryingClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return time.Second
	}
	return jsonrpc2.NewCustomHTTPClient(kurtosisUrl, retryingClient.StandardClient())
}
