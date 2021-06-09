/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_mounting_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis-libs/golang/testsuite/services_impl/nginx_static"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	fileServerServiceId services.ServiceID = "file-server"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxRetries = 15

	testFilesArtifactId  services.FilesArtifactID = "test-files-artifact"
	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	// Filenames & contents for the files stored in the files artifact
	file1Filename = "file1.txt"
	file2Filename = "file2.txt"

	expectedFile1Contents = "file1"
	expectedFile2Contents = "file2"
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
	configFactory := nginx_static.NewNginxStaticContainerConfigFactory(testFilesArtifactId)
	_, hostPortBindings, availabilityChecker, err := networkCtx.AddService(fileServerServiceId, configFactory)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}
	if err := availabilityChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxRetries); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the file server service to start")
	}
	logrus.Infof("Added file server service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (f FilesArtifactMountingTest) Run(network networks.Network) error {
	// Only necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	uncastedService, err := castedNetwork.GetService(fileServerServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving the fileserver service")
	}

	// Only necessary because Go doesn't have generics
	castedService, castErrOccurred := uncastedService.(*nginx_static.NginxStaticService)
	if castErrOccurred {
		return stacktrace.Propagate(err, "An error occurred casting the file server service API")
	}

	file1Contents, err := castedService.GetFileContents(file1Filename)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting file 1's contents")
	}
	if file1Contents != expectedFile1Contents {
		return stacktrace.NewError("Actual file 1 contents '%v' != expected file 1 contents '%v'",
			file1Contents,
			expectedFile1Contents,
		)
	}

	file2Contents, err := castedService.GetFileContents(file2Filename)
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
