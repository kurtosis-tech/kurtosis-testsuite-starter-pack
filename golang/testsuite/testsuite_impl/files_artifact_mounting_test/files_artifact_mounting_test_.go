/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_mounting_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/nginx_static"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const (
	fileServerServiceId services.ServiceID = "file-server"

	waitForStartupTimeBetweenPolls = 1000
	waitForStartupMaxRetries = 15
	waitInitialDelaySeconds = 0

	testFilesArtifactId  services.FilesArtifactID = "test-files-artifact"
	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	// Filenames & contents for the files stored in the files artifact
	file1Filename = "file1.txt"
	file2Filename = "file2.txt"

	expectedFile1Contents = "file1\n"
	expectedFile2Contents = "file2\n"
)

type FilesArtifactMountingTest struct {}

func (f FilesArtifactMountingTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(
		60,
	).WithRunTimeoutSeconds(
		60,
	).WithFilesArtifactUrls(
		map[services.FilesArtifactID]string{
			testFilesArtifactId: testFilesArtifactUrl,
		},
	)
}

func (f FilesArtifactMountingTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {

	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		"flashspys/nginx-static",
		"/test-volume",
	).WithUsedPorts(
		map[string]bool{"80": true},
	).WithFilesArtifacts(map[services.FilesArtifactID]string{
		"test-files-artifact": "/static",
	}).Build()

	generateRunConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		return services.NewContainerRunConfigBuilder().Build(), nil
	}

	_, hostPortBindings, err := networkCtx.AddService(fileServerServiceId, containerCreationConfig, generateRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}

	if err := networkCtx.WaitForEndpointAvailability(fileServerServiceId, nginx_static.ListenPort, file1Filename, waitInitialDelaySeconds, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls, ""); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the file server service to become available")
	}

	logrus.Infof("Added file server service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (f FilesArtifactMountingTest) Run(uncastedNetwork networks.Network) error {
	// Necessary because Go doesn't have generics
	network := uncastedNetwork.(*networks.NetworkContext)

	fileServerServiceContext, err := network.GetServiceContext(fileServerServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service context with ID '%v'", fileServerServiceId)
	}

	file1Contents, err := getFileContents(fileServerServiceContext.GetIPAddress(), nginx_static.ListenPort, file1Filename)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting file 1's contents")
	}
	if file1Contents != expectedFile1Contents {
		return stacktrace.NewError("Actual file 1 contents '%v' != expected file 1 contents '%v'",
			file1Contents,
			expectedFile1Contents,
		)
	}

	file2Contents, err := getFileContents(fileServerServiceContext.GetIPAddress(), nginx_static.ListenPort, file2Filename)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting file 2's contents")
	}
	if file2Contents != expectedFile2Contents {
		return stacktrace.NewError("Actual file 2 contents '%v' != expected file 2 contents '%v'",
			file2Contents,
			expectedFile2Contents,
		)
	}
	return nil
}

func getFileContents(ipAddress string, port uint32, filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, port, filename))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", filename)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}
