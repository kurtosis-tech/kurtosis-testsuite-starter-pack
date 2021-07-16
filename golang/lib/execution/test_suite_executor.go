package execution

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/rpc_api/rpc_api_consts"
	"github.com/kurtosis-tech/minimal-grpc-server/server"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"time"
)

const (
	grpcServerStopGracePeriod = 5 * time.Second
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

	var apiContainerService kurtosis_core_rpc_api_bindings.ApiContainerServiceClient = nil
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

		apiContainerService = kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)
	}

	testsuiteService := NewTestSuiteService(suite, apiContainerService)
	testsuiteServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		bindings.RegisterTestSuiteServiceServer(grpcServer, testsuiteService)
	}

	testsuiteServer := server.NewMinimalGRPCServer(
		rpc_api_consts.ListenPort,
		rpc_api_consts.ListenProtocol,
		grpcServerStopGracePeriod,
		[]func(desc *grpc.Server) {
			testsuiteServiceRegistrationFunc,
		},
	)
	if err := testsuiteServer.Run(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the testsuite server")
	}

	return nil
}
