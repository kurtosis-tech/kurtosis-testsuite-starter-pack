Kurtosis Libs
==================
This repo contains:

1. Libraries in various languages for interacting with Kurtosis Core, which are used to write testsuites
1. Example implementations of testsuites in each langauge
1. Infrastructure for bootstrapping a new testsuite, that you can use to create your own customized testsuite

### Testsuite Quickstart
Prerequisites:
* A [Kurtosis user account](https://www.kurtosistech.com/sign-up)
* `git` installed on your machine
* `docker` installed on your machine

Quickstart steps:
1. Clone this repo's `master` branch: `git clone --single-branch --branch master $THIS_REPO_URL`
1. View [the supported languages](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/supported-languages.txt) and choose the language you'd like your testsuite in
1. Run `bootstrap/bootstrap.sh $DESIRED_LANG /path/to/output/testsuite/dir` and follow the instructions

If you see error messages after running your new testsuite, check out [the guide for debugging failed tests](./debugging-failed-tests.md) which contains solutions to common issues. If this still doesn't resolve your issue, feel free to ask for help in [the Kurtosis Discord server](https://discord.gg/6Jjp9c89z9).

If all tests are passing, you can [proceed to customizing your testsuite](https://docs.kurtosistech.com/testsuite-customization.html).

### Developing Libraries
Prerequisites:
* `protoc` installed (can be installed on Mac with `brew install protobuf`)
* The Golang extension to `protoc` installed (can be installed on Mac with `brew install protoc-gen-go`)
* The Golang gRPC extension to `protoc` installed (can be installed on Mac with `brew install protoc-gen-go-grpc`)
* [rust-protobuf-binding-generator](https://github.com/kurtosis-tech/rust-protobuf-binding-generator) installed

_NOTE: One day we want to push all the protobuf binding into Docker, so that the output doesn't depend on the developer's machine; see [this issue](https://github.com/kurtosis-tech/kurtosis-libs/issues/22) for more details_

Each library needs to talk with Kurtosis Core, and the Kurtosis Core API is defined via Protobuf. Rather than storing the Protobufs in Git submodules (which add significant complexity), the `.proto` files are simply copied from the relevant version of Kurtosis Core. In the future, we can move to a more productized solution.

To regenerate the bindings corresponding to the Protobuf files, use the `scripts/regenerate-protobuf-output.sh` script.
