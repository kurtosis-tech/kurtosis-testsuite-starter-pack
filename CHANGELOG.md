## TBD
* Refactor directory structure and `regenerate-protobuf-output.sh` script to support multiple languages

## 1.7.1
* Do a better job grabbing the name of the current Git ref
* Remove scary bootstrapping message with a more reasonable verification

## 1.7.0
* Make README point to the quickstart docs, rather than duplicating them
* Upgraded to use the new Protobuf-based APIs of Kurtosis Core 1.7.0
* Replaced the Kurtosis `Client` with `TestExecutor` and `TestExecutorConfigurator` to allow users to set log level, parse custom params, and initialize their testsuite without needing to modify `main.go` or the `Dockerfile`
* Switched to receiving custom testsuite params via the `CUSTOM_PARAMS_JSON` Docker environment variable
* Pushed a large amount of logic from that was in Kurtosis Go (particularly `NetworkContext`) into Kurtosis Core, so that the Go library is just a thing wrapper over Kurtosis Core
    * Added logic (transparent to the end user) inside TestExecutor for running the test execution or metadata serialization flows
* Removed a ton of now-unnecessary Docker environment variables:

## 1.6.1
* Upgrade to `kurtosis.sh` script that will pull the latest Docker Kurtosis Core images automatically
* Fail CI if we detect the string `ERRO`, to catch problems that don't get propagated to the exit code (e.g. not printing the testsuite container logs)
* Don't break on empty `${@}` in build_and_run (regression introduced when switching to `kurtosis.sh`)
* Add a `IS_KURTOSIS_CORE_DEV_MODE` custom param to the testsuite, so that we can:
    1. Run extra tests when testing Kurtosis Core but
    2. Disable those extra tests after a user has bootstrapped, since it will exceed their free trial test limit
* Make sure `.dockerignore` gets created after bootstrapping

## 1.6.0
* Use Kurtosis Core v1.6.0
* Implement API changes to allow users to mount external artifacts inside Kurtosis service containers
    * The Kurtosis client now must be instantiated with `NewKurtosisClient`
    * Added `FilesArtifactUrls` property to `TestConfiguration` to declare files artifacts
    * Added `GetFilesArtifactMountpoints` to `DockerContainerInitializer` to use files artifacts in a service
* Added `FilesArtifactMountingTest` to test the new external artifact-mounting functionality

## 1.5.0
* Add a `.dockerignore` file, and a check in `build_and_run.sh` to ensure it exists
* Add the `Service.GetServiceID` method
* Renamed `DockerContainerInitializer.GetServiceFromIp` -> `GetService`, and passed in the `ServiceID` as a new first argument
    * All `Service` implementations should have their constructors modified to store this new argument
* Implemented the ability to partition test networks! This brought several changes:
    * Upgraded to Kurtosis Core 1.5
    * Added a `GetTestConfiguration` function to the `Test` interface, which allows tests to declare certain types of functionality (like network partitioning)
    * Added `NetworkPartitionTest` to test the new network partitioning functionality
    * Made `NetworkContext` thread-safe
    * Add tests for `RepartitionerBuilder` actions
    * Added extra testing inside `NetworkPartitionTest` to ensure that a node that gets added to a partition receives the correct blocking
* Remove the HTTP client retrying from the JSON RPC client, because it can obscure errors like panics in the Kurtosis API and lead to red herring errors as it replays the call when the problem was the 
* Added the ability to mount external files into a service container:
    * Added a new property, `FilesArtifactUrsl`, to `TestConfiguration` for defining files artifact URLs
    * Add a new method, `GetFilesArtifactMountpoints`, to `DockerContainerInitializer` for defining where to mount external files artifacts
    * Add `FilesArtifactTest` to test pulling down a files artifact, mounting it inside a service, and using those files

## 1.4.1
* Point all old `kurtosis-docs` references to `docs.kurtosistech.com`
* Switch `build_and_run.sh` to use `kurtosis.sh`
* Upgrade to Kurtosis Core 1.4
* Reduce the size of the testsuite image by using the `golang` image only for building, and then `alpine` for execution; this results in a reduction of 325 MB -> 14 MB

## 1.4.0
* BREAKING: Moved `ServiceID` from the `networks` package to the `services` package
* Add a more explanatory help message to `build_and_run`
* After calling `bootstrap.sh`, ensure the volume is named based off the name of the user's Docker image
* Update the example testsuite to use the Kurtosis-developed example API service and example datastore service, to show dependencies and file generation

## 1.3.0
* Bump kurtosis-core-channel to 1.2.0
* Heavily refactored the client architecture to make it much less confusing to define testsuite infrastructure:
    * The notion of `dependencies` that showed up in several places (e.g. `ServiceInitializerCore.GetStartCommand`, `ServiceAvailabilityCheckerCore.IsServiceUp`, etc) have been removed due to being too confusing
    * Services: 
        * The `Service` interface (which used to be a confusing marker interface) has now received `GetIPAddress` and `IsAvailable` to more accurately reflect what a user expects a service to be
        * `ServiceInitializerCore`, `ServiceInitializer`, and `ServiceAvailabilityCheckerCore` have been removed to cut down on the number of components users need to write & remember
        * `ServiceInitializerCore`'s functionality has been subsumed by a new interface, `DockerContainerInitializer`, to more accurately reflect what its purpose
        * `ServiceAvailabilityChecker` renamed to `AvailabilityChecker` to make it easier to say & type
    * Networks: 
        * `ServiceNetwork` has been renamed to `NetworkContext` to more accurately reflect its purpose
        * `NetworkContext.AddService` has been made easier to work with by directly returning the `Service` that gets added (rather than a `ServiceNode` package object)
        * Test networks are no longer created in two separate configuration-then-instantiation phases, and are simply instantiated directly in the new `Test.Setup` method
        * The notion of "service configuration" that was used during the network configuration phase has been removed, now that networks are instantiated directly in `Test.Setup`
        * `ServiceNetworkBuilder` has been removed
        * `NetworkLoader` has been removed
    * Testsuite:
        * `Test.GetSetupBuffer` has been renamed to `GetSetupTeardownBuffer` to more accurately reflect its purpose
        * The `Test.GetNetworkLoader` method has been replaced with `Test.Setup(NetworkContext) Network` to simplify network instantiation and more closely match other test frameworks
            * The `Network` return type is still `interface{}`, so users can return `NetworkContext` directly or wrap it in a more test-friendly custom object
        * Kurtosis no longer controls network availability-checking, which lets users do it however they please in `Test.Setup` (e.g. start all services in parallel then wait for them to come up, start them in serial, skip it entirely, etc.)
            * An `AvailabilityChecker` is still returned by `NetworkContext.AddService`, so waiting on a service is still simple
* Disable logging from the RetryingHTTPClient inside `KurtosisService`, as the output isn't useful (and can be unnecessarily alarming, when a request fails)
* Remove the `FixedSizeNginxNetwork` from the example implementation, to demonstrate a simpler `Test.Setup` usage without a custom `Network`

## 1.2.0
* Remove socket in favor of `ExampleService.GetIpAddress` and `ExapleService.GetPort` methods
* Remove TODO on allowing non-TCP ports
* Removed the `example_` prefix to make bootstrapping even easier (users need only modify the existing suite, no need to remove the `example_` prefix)
* Support UDP ports

## 1.1.1
* Remove log filepath (which is no longer needed now that Kurtosis core reads Docker logs directly)
* Switch to using [our forked version of action-comment-run](https://github.com/mieubrisse/actions-comment-run) that allows user whitelisting
* Bump kurtosis-core to 1.1.0
* Make the requests to the Kurtosis API container retry every second, with 10s retry maximum for normal operations (e.g. add/remove services) and 60s retry maximum for test suite registration
* Update the version of the `actions-comment-run` Github Action which allows for running CI on untrusted PRs, to match the advice we give in the "Running In CI" instructions

## 1.1.0
* Add Apache license
* Fix changelog check in CircleCI 
* Cleaning TODOs 
* Added a README pointing users to the main Kurtosis docs
* Cleaned up `build_and_run.sh` with lessons learned from upgrading the Avalanche test suite to Kurtosis 1.0
* Explain nil start command for the example impl
* Added a new bootstrapping process for creating Kurtosis Go testsuites from scratch
* Add [the comment-run Github Action](https://github.com/nwtgck/actions-comment-run/tree/20297f070391450752be7ac1ebd454fb53f62795#pr-merge-preview) to the repository in order to set up [a workaround for Github not passing secrets to untrusted PRs](https://github.community/t/secrets-for-prs-who-are-not-collaborators/17712), which would prevent auth'd Kurtosis from running
* Simplified the bootstrapping process quite a bit
* In addition to building `develop` and `master` images, build `X.Y.Z` tag images
* Cleaned up an over-aggressive check that was causing testsuite log-listing to fail
* When no arguments are provided to `build_and_run.sh`, the script errors
* In CircleCI config, don't run the `validate` workflow on `develop` and `master` (because they should already be validated by PR merge)

## 1.0.0
* Created example test suite to validate that the client library work
* Bugfix in volume-writing location, and force pretty formatting on written logs
* Made the existing test actually query the node it created
* Added another test to demonstrate an initial network setup
* Adding copyright headers
* Renamed tests to have more descriptive names
* When asked about test suite data, send back a JSON of test suite metadata (rather than just a list of test names)
* Made log level configurable
* Add CircleCI
* Upgraded example Go implementation to show the use of custom environment variables
* Build a Docker image on each merge to the develop branch
* Accept a new Docker parameter, `SERVICES_RELATIVE_DIRPATH`, for the location (relative to the suite execution volume root) where file IO for the services created during test execution
* Consolidate all the scripts into `build_and_run.sh` which will actually run the test suite for testing purposes
* Switch to `master` release track from Kurtosis core
