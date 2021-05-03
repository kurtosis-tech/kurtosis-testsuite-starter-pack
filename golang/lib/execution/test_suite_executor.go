package execution

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	grpcServerStopGracePeriod = 10 * time.Second
)

type TestSuiteExecutor struct {
	kurtosisApiSocket string  // Can be empty if the testsuite is in metadata-providing mode
	logLevelStr string
	paramsJsonStr string
	configurator TestSuiteConfigurator
}

func NewTestSuiteExecutor(kurtosisApiSocket string, logLevelStr string, paramsJsonStr string, configurator TestSuiteConfigurator) *TestSuiteExecutor {
	return &TestSuiteExecutor{kurtosisApiSocket: kurtosisApiSocket, logLevelStr: logLevelStr, paramsJsonStr: paramsJsonStr, configurator: configurator}
}

func (executor TestSuiteExecutor) Run() error {
	if err := executor.configurator.SetLogLevel(executor.logLevelStr); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting the loglevel before running the testsuite")
	}

	suite, err := executor.configurator.ParseParamsAndCreateSuite(executor.paramsJsonStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the suite params JSON and creating the testsuite")
	}

	var apiContainerService core_api_bindings.ApiContainerServiceClient = nil
	if executor.kurtosisApiSocket != "" {
		// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
		conn, err := grpc.Dial(executor.kurtosisApiSocket, grpc.WithInsecure())
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred creating a connection to the Kurtosis API server at '%v'",
				executor.kurtosisApiSocket,
			)
		}
		defer conn.Close()

		apiContainerService = core_api_bindings.NewApiContainerServiceClient(conn)
	}

	testsuiteService := NewTestSuiteService(suite, apiContainerService)

	// TODO all the code below here is almost entirely duplicated with KUrtosis Core - extract as a library!
	grpcServer := grpc.NewServer()

	bindings.RegisterTestSuiteServiceServer(grpcServer, testsuiteService)

	listenAddressStr := fmt.Sprintf(":%v", rpc_api_consts.ListenPort)
	listener, err := net.Listen(rpc_api_consts.ListenProtocol, listenAddressStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the listener on %v/%v",
			rpc_api_consts.ListenProtocol,
			listenAddressStr,
		)
	}

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	grpcServerResultChan := make(chan error)

	go func() {
		var resultErr error = nil
		if err := grpcServer.Serve(listener); err != nil {
			resultErr = stacktrace.Propagate(err, "The gRPC server exited with an error")
		}
		grpcServerResultChan <- resultErr
	}()

	// Wait until we get a shutdown signal
	<- termSignalChan

	serverStoppedChan := make(chan interface{})
	go func() {
		grpcServer.GracefulStop()
		serverStoppedChan <- nil
	}()
	select {
	case <- serverStoppedChan:
		logrus.Info("gRPC server has exited gracefully")
	case <- time.After(grpcServerStopGracePeriod):
		logrus.Warnf("gRPC server failed to stop gracefully after %v; hard-stopping now...", grpcServerStopGracePeriod)
		grpcServer.Stop()
		logrus.Info("gRPC server was forcefully stopped")
	}
	if err := <- grpcServerResultChan; err != nil {
		// Technically this doesn't need to be an error, but we make it so to fail loudly
		return stacktrace.Propagate(err, "gRPC server returned an error after it was done serving")
	}

	return nil

}
