package local_static_file_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
)

type LocalStaticFileTest struct {

}

func (l LocalStaticFileTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	panic("implement me")
}

func (l LocalStaticFileTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	panic("implement me")
}

func (l LocalStaticFileTest) Run(network networks.Network) error {
	panic("implement me")
}
