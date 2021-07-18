package lambda_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
)

const (
	testLambdaImage = "mieubrisse/datastore-army-lambda"

	datastoreArmyLambdaId modules.LambdaID = "datastore-army"
)

type LambdaTest struct {}

func (l LambdaTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (l LambdaTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	lambdaCtx, err := networkCtx.LoadLambda(datastoreArmyLambdaId, testLambdaImage, "{}")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore army Lambda")
	}

}

func (l LambdaTest) Run(network networks.Network) error {
	panic("implement me")
}

