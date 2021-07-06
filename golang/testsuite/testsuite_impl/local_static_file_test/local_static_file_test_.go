package local_static_file_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
)

type LocalStaticFileTest struct {

}

func (l LocalStaticFileTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (l LocalStaticFileTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	return nil, stacktrace.NewError("Not implemented")
}

func (l LocalStaticFileTest) Run(network networks.Network) error {
	return stacktrace.NewError("Not implmeented")
}
