_See [here](./versioning-and-upgrading.md) for information about versioning and upgrading_

# TBD
### Changes
* Added several clarifications to the bootstrap onboarding process after a user research session

# 1.24.3
### Changes
* Harden `validate-all-bootstraps` error-checking

### Fixes
* Fix an error with bootstrap script where it would pull an old version of `kurtosis-libs` if a new version had just been released
* Fix an error in `validate-all-bootstraps` script

# 1.24.2
### Changes
* Made the `git clone` command in the quickstart copy-pasteable by filling in `$THIS_REPO_URL`
* Make the command to run the new testsuite after a bootstrap use `parallelism=1`, so the user can get immediate feedback that things are running

### Features
* Added a CI check to verify that all the links in Markdown files point to the correct location

### Fixes
* Upgraded `kurtosis.sh` to latest version to fix bug with UUID uppercasing failing in Zshell
* Fixed an occasional failure due to the initializer trying to connect to the testsuite container before it's ready
* Fixed an issue with the bootstrap helptext missing a flag
* Made bootstrap point back to the quickstart flow
* Simplified the quickstart flow to include error-checking and proceeding to testsuite customization
* Added clarification to the previously-terse descriptions of the bootstrap script args
* Clarified that you shouldn't prefix `https://` when specifying the Go module
* Fixed indentation bug in bootstrap script

# 1.24.1
### Fixes
* Fixed bug in bootstrapper script when providing a relative output directory

# 1.24.0
### Changes
* Upgraded Kurtosis Core to 1.14, which contains backend changes in preparation for some new upcoming features
* Dropped support for Rust because it's a large burden to maintain and not currently being used
    * NOTE: This can be resurrected quite quickly if needed - if you need testsuites in Rust, please get in touch!

### Breaking Changes
* Dropped support for Rust
* Upgraded to Kurtosis Core 1.14 (requires downloading correct scripts from the [dists page](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/))
* For Go testsuites, `TestSuiteExecutor.Run` no longer takes in a `Context` object

# 1.23.0
### Features
* Upgraded Kurtosis Core to 1.13, which gives descriptive names to Docker containers started by Kurtosis

### Changes
* Refactored the example `Network` implementation, `TestNetwork`, to provide a better example by more closely aligning with what we've seen in the real world - a single `Setup` method intended for use inside `Test.Setup`, and several getters to retrieve the values created during setup
* Upped the default `Test.Setup` and `Test.Run` timeouts from 60s to 180s

### Fixes
* Corrected broken link in README

### Breaking Changes
* Upgraded to Kurtosis Core 1.13 (requires downloading correct scripts from the [dists page](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/))

# 1.22.0
### Features
* The `kurtosis.sh` script now takes in a `--debug` argument that will, among other things, instruct Kurtosis to bind all the ports used by any service in `ContainerCreationConfig.usedPorts` to a port on the user's local machine, so the user can make queries to the services inside Kurtosis directly
    * Full information is available by passing in the `--help` flag

### Breaking Changes
* Upgraded Kurtosis Core to version 1.12 (requires downloading correct scripts from the [dists page](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/))
* `NetworkContext.addService` and `NetworkContext.addServiceToPartition` now return an extra argument, which contains the interface IP & port on the Docker host machine where the service's ports have been bound

# 1.21.1
### Fixes
* Bootstrap-validating CircleCI job was running before the Rust libs & Docker image were pushed and failing, so added a dependency on them to prevent this

# 1.21.0
### Features
* The specifications for starting a container are now provided via the `ContainerConfigFactory` interface (rather than `DockerContainerInitializer`), with the actual container config created via `ContainerCreationConfigBuilder` and `ContainerRunConfigBuilder`. This was done so that:
    * Optional features are only specified if needed (e.g. many containers won't need files artifacts, so users shouldn't need to fill out a `getFilesArtifact` function like they had to with `DockerContainerInitializer`)
    * New features won't cause an API break (adding a new feature usually meant a new function in the `DockerContainerInitializer` interface)
* Added docs for `ContainerConfigFactory` to [the Kurtosis Libs documentation](https://docs.kurtosistech.com/kurtosis-libs/lib-documentation)

### Fixes
* Fixed issue with `validate-all-bootstraps.sh` where it would clobber your Git `user.email` and `user.name` if you run it on a machine where these are already set up

### Breaking Changes
* The `DockerContainerInitializer` interface has been replaced with the `ContainerConfigFactory` interface
* The functionality in the `getDockerImage`, `getTestVolumeMountpoint`, `getUsedPorts`, `getService`, `getFilesToGenerate`, `initializeGeneratedFiles`, and `getFilesArtifactMountpoints` functions have been moved to `ContainerCreationConfigBuilder`, which should be used in `ContainerConfigFactory.getCreationConfig`
* The functionality in the `getStartCommandOverrides` and `getEnvironmentVariableOverrides` functions have been moved to `ContainerRunConfigBuilder`, which should be used in `ContainerConfigFactory.getRunConfig`
* `NetworkContext.addService` and `NetworkContext.addServiceToPartition` take in a `ContainerConfigFactory` argument, rather than `DockerContainerInitializer`


# 1.20.0
### Features
* Added a `ServiceContext.generateFiles` method that can be used to generate files inside a testsuite container on the fly, with docs available on [docs.kurtosistech.com](https://docs.kurtosistech.com/)

### Breaking Changes
* Upgraded to Kurtosis Core 1.11 (requires downloading correct scripts from the [dists page](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/))

# 1.19.6
### Features
* Run the `validate-all-bootstraps` CI script using simulated trial user creds, to simulate a new user bootstrapping as closely as possible

# 1.19.5
### Fixes
* Fixed issue with bootstrap validation failing due to Kurtosis client ID & secret not being used

# 1.19.4
### Fixes
* Fixed issue with bootstrap validation failing due to Git not being initialized

# 1.19.3
### Fixes
* Fix issue where bootstrap scripts would fail on Linux

# 1.19.2
### Changes
* Added `set -x` to Go `prep-new-repo` script to debug why it's failing the bootstrap check

# 1.19.1
### Features
* Added a CI check on merge-to-master that validates that all language bootstraps work

# 1.19.0
### Changes
* Made test configuration easier to define via a `Test.configure` method that allows users to set test configuration using a `TestConfigurationBuilder` object
    * This has the added benefit that test configurations which get added in the future won't cause breaking changes
* Centralized the test setup & run timeouts to be configured via `TestConfigurationBuilder` in the `Test.configure` method

### Breaking Changes
* `Test.getTestConfiguration` was removed, and replaced with `Test.configure` which consumes a `TestConfigurationBuilder` and is how all test configuration will be defined going forward
* `Test.getSetupTimeout` was removed, replaced with `TestConfigurationBuilder.withSetupTimeoutSeconds`, set in `Test.configure`
* `Test.getExecutionTimeout` was removed, replaced with `TestConfigurationBuilder.withRunTimeoutSeconds`, set in `Test.configure`

# 1.18.0
### Features
* Made this changelog available as an HTML webpage

### Fixes
* Bootstrapping a new Rust testsuite now no longer keeps the Kurtosis Lib version in Cargo.toml, and instead gets `version = "0.1.0"`

### Changes
* The `TestContext` interface was removed, as it duplicated functionality with the return type of the `Test.run` function

### Breaking Changes
* The `Test.run` method no longer takes in a `TestContext` argument
    * To fail the test, your tests should now simply return an error
* The Go lib's `Test.Run` method now returns an `error` type for indicating that an error occurred while running the test

# 1.17.0
### Features
* Added unit test to Rust `AvailabilityChecker` (ported over from Go)
* Allow lib users to set the Docker environment variables when starting a service

### Breaking Changes
* Addded `DockerContainerInitializer.getEnvironmentVariableOverrides` for setting environment variables when launching services

# 1.16.0
### Changes
* Added type aliases `ServiceId` and `PartitionId` to Rust library and changed function signatures to match

### Breaking Changes
* Modified several Rust functions slightly to take in `ServiceId` and `PartitionId` rather than `String`

# 1.15.1
### Changes
* Replaced all docstring comments on all lib classes/methods with a link to https://docs.kurtosistech.com/kurtosis-libs/lib-documentation , to centralize documentation

### Fixes
* Fix several bugs with the lib documentation

# 1.15.0
### Features
* Added language-agnostic documentation for each class & function, which gets published to https://docs.kurtosistech.com

### Fixes
* Fixed Rust's `DockerContainerInitializer` parameterized with `S: Service`, but `getService` was returning `Box<dyn Service>` (so an initializer could theoretically return a service not matching the initializer)

### Breaking Changes
* Rust `DockerContainerIniitalizer.getService` now returns `Box<S>` rather than `Box<dyn Service>`

# 1.14.3
### Features
* Add `docs` directory, for publishing with Github Pages

# 1.14.2
### Fixes
* Fixes the issue where the Rust testsuite would need to be built twice whenever Cargo.toml changed due to `cargo` not rebuilding the binary if the source code hasn't changed (see https://github.com/emk/rust-musl-builder/issues/101 )

# 1.14.1
### Features
* Switched the Rust library's `ServiceContext.exec_command` to use `&self` rather than `&mut self`

# 1.14.0
### Fixes
* Added an error log message when the Rust testsuite errors so the exact timestamp of the failure is visible, which brings it to parity with the Go testsuite
* Simplified the way services are created, by making `DockerContainerInitializer`s construct `Service` instances directly, rather than via a service wrapping function
* Removes the intermixing of `tokio` and `block_on`, which causes deadlocks, in favor of pure `tokio`

### Breaking Changes
* `DockerContainerInitializer.getServiceWrappingFunction` name & signature changed to `getService(ServiceContext) -> Service`

# 1.13.0
### Features
* Docker exec commands now have their log output available

### Breaking Changes
* `ServiceContext.ExecCommand` now returns an extra argument, the bytes of the log output from the exec command

# 1.12.1
### Features
* Add extra debugging logic when going through the setup process during test execution flow

# 1.12.0
### Features
* Users can now optionally override a Docker image's `ENTRYPOINT` directive

### Breaking Changes
* Renamed `DockerContainerInitializer.getStartCommand` -> `getStartCommandOverrides` to better reflect that both `ENTRYPOINT` and `CMD` directives can be overridden
* `DockerContainerInitializer.getStartCommandOverrides` now returns two string arrays - one for overriding a Docker image's `ENTRYPOINT`, and one for overriding its `CMD`

# 1.11.0
_NOTE: This changelog entry abandons the old KeepAChangelog subheadings because they didn't do a good job highlighting the important bits - features, fixes, and breaking changes_

### Features
* Bootstrapping a Rust testsuite now slots in a user-provided package name to `testsuite/Cargo.toml`
* Encapsulated the logic for getting a service's ID and IP address in a new `ServiceContext` object
* Added support to allow testsuites to `docker exec` commands inside their service containers via `ServiceContext`, to allow for things like assertions on the internal state of the container
* Added an `ExecCommandTest` to the Go testsuite, for regression-testing the `ServiceContext` Docker exec

### Fixes
* Removed an internal-only comment in Rust's `testsuite/Cargo.toml` that was getting incorrectly propagated to bootstrapped testsuites

### Breaking Changes
* Added a `ServiceContext` object to represent a service container, with the intention that user implementations of the `Service` interface store the `ServiceContext`
* `DockerContainerInitializer.getServiceWrappingFunction` now takes in a `ServiceContext` arg, rather than the IP address and service ID separately
* Removed the `GetServiceID` and `GetIPAddress` functions from the `Service` interface (now handled by `ServiceContext`)

# 1.10.5
### Fixed
* Removed the Rust 30-second timeout on requests to the API container, which could get tripped on long requests (e.g. if the API container has to download a big Docker image). The test setup/execution timeouts serve as a backstop regardless, to prevent forever-hung requests from running forever.

# 1.10.4
### Added
* Rust testsuite bootstrapping

### Changed
* Fixed `release.sh` so it doesn't require taking the version-to-release three times

# 1.10.3
### Changed
* Modified CircleCI config to _only_ push example testsuite Docker images for `X.Y.Z`-tagged commits (no longer Docker images with `develop` or `master` tags)

# 1.10.2
### Changed
* Fixed more `kurtosis-rust-lib` Cargo.toml errors preventing publishing

# 1.10.1
### Added
* Extra metadata to `kurtosis-rust-lib`'s Cargo.toml

# 1.10.0
### Added
* Rust example testsuite to the set of testsuites that get checked in the CircleCI config
* A `release.sh` script which will run the process of cutting a new release for this repo (necessary because we have to update the `version` key in Rust's Cargo.toml files)

### Changed
* Refactored CircleCI config to support validating multiple testsuites & pushing multiple example testsuite Docker images
* Modified CircleCI building to skip building a language's testsuite if a) no shared code has changed and b) the language's directory doesn't have any changes
* Renamed `DockerContainerInitializer.GetFilesToMount` to `GetFilesToGenerate`, and `InitializeMountedFiles` to `InitializeGeneratedFiles`, in both Rust and Go

# 1.9.0
### Added
* Rust client library

### Changed
* Replaced the `DockerContainerInitializer`'s `GetService` function with a `GetServiceWrappingFunc` factory-getting function so that `NetworkContext` can manufacture `Service` instances upon `GetService` request
* Made Golang `NetworkContext.Repartition` take in the new network state directly, rather than going through the `Repartitioner`/`RepartitionerBuilder`
* Renamed `GetSetupTeardownBuffer` -> `GetSetupTimeout` on the `Test` interface

### Removed
* `Repartitioner` and `RepartitionerBuilder`, which added complexity for no gain

# 1.8.1
### Fixed
* `kurtosis.sh` issue where passing in `--help` would result in an error

### Added
* Additional quickstarting instructions in README
* Added a helptext to the generated `build-and-run.sh` after bootstrapping

### Changed
* Moved `Dockerfile` inside the `testsuite` repo of each language once again
* Modified API container API to control test setup and execution timeouts in Kurtosis Core instead of kurtosis libs

# 1.8.0
* Refactor directory structure and `regenerate-protobuf-output.sh` script to support multiple languages
* Fixing CircleCI to work with new `kurtosis-libs` repo
* Fix `build_and_run` to work with new repo
* Push `go test` into the golang Dockerfile
* Refactor `build_and_run.sh` so that it can be completely Kurtosis-controlled:
    * Rename `build_and_run.sh` to `build-and-run-core.sh` and put it in `scripts` directory at the root of this repo
    * Place `build-and-run.sh` inside the `scripts` directory of the Golang subdirectory, that calls down to `build-and-run-core.sh`
* Fix bootstrapping to support multiple languages by

# 1.7.1
* Do a better job grabbing the name of the current Git ref
* Remove scary bootstrapping message with a more reasonable verification

# 1.7.0
* Make README point to the quickstart docs, rather than duplicating them
* Upgraded to use the new Protobuf-based APIs of Kurtosis Core 1.7.0
* Replaced the Kurtosis `Client` with `TestExecutor` and `TestExecutorConfigurator` to allow users to set log level, parse custom params, and initialize their testsuite without needing to modify `main.go` or the `Dockerfile`
* Switched to receiving custom testsuite params via the `CUSTOM_PARAMS_JSON` Docker environment variable
* Pushed a large amount of logic from that was in Kurtosis Go (particularly `NetworkContext`) into Kurtosis Core, so that the Go library is just a thing wrapper over Kurtosis Core
    * Added logic (transparent to the end user) inside TestExecutor for running the test execution or metadata serialization flows
* Removed a ton of now-unnecessary Docker environment variables:

# 1.6.1
* Upgrade to `kurtosis.sh` script that will pull the latest Docker Kurtosis Core images automatically
* Fail CI if we detect the string `ERRO`, to catch problems that don't get propagated to the exit code (e.g. not printing the testsuite container logs)
* Don't break on empty `${@}` in build_and_run (regression introduced when switching to `kurtosis.sh`)
* Add a `IS_KURTOSIS_CORE_DEV_MODE` custom param to the testsuite, so that we can:
    1. Run extra tests when testing Kurtosis Core but
    2. Disable those extra tests after a user has bootstrapped, since it will exceed their free trial test limit
* Make sure `.dockerignore` gets created after bootstrapping

# 1.6.0
* Use Kurtosis Core v1.6.0
* Implement API changes to allow users to mount external artifacts inside Kurtosis service containers
    * The Kurtosis client now must be instantiated with `NewKurtosisClient`
    * Added `FilesArtifactUrls` property to `TestConfiguration` to declare files artifacts
    * Added `GetFilesArtifactMountpoints` to `DockerContainerInitializer` to use files artifacts in a service
* Added `FilesArtifactMountingTest` to test the new external artifact-mounting functionality

# 1.5.0
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

# 1.4.1
* Point all old `kurtosis-docs` references to `docs.kurtosistech.com`
* Switch `build_and_run.sh` to use `kurtosis.sh`
* Upgrade to Kurtosis Core 1.4
* Reduce the size of the testsuite image by using the `golang` image only for building, and then `alpine` for execution; this results in a reduction of 325 MB -> 14 MB

# 1.4.0
* BREAKING: Moved `ServiceID` from the `networks` package to the `services` package
* Add a more explanatory help message to `build_and_run`
* After calling `bootstrap.sh`, ensure the volume is named based off the name of the user's Docker image
* Update the example testsuite to use the Kurtosis-developed example API service and example datastore service, to show dependencies and file generation

# 1.3.0
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

# 1.2.0
* Remove socket in favor of `ExampleService.GetIpAddress` and `ExapleService.GetPort` methods
* Remove TODO on allowing non-TCP ports
* Removed the `example_` prefix to make bootstrapping even easier (users need only modify the existing suite, no need to remove the `example_` prefix)
* Support UDP ports

# 1.1.1
* Remove log filepath (which is no longer needed now that Kurtosis core reads Docker logs directly)
* Switch to using [our forked version of action-comment-run](https://github.com/mieubrisse/actions-comment-run) that allows user whitelisting
* Bump kurtosis-core to 1.1.0
* Make the requests to the Kurtosis API container retry every second, with 10s retry maximum for normal operations (e.g. add/remove services) and 60s retry maximum for test suite registration
* Update the version of the `actions-comment-run` Github Action which allows for running CI on untrusted PRs, to match the advice we give in the "Running In CI" instructions

# 1.1.0
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

# 1.0.0
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
