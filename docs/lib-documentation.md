Test SDK Documentation
======================
The Kurtosis testing SDK exists in multiple languages and maintaining in-code comments for each of these is prohibitively expensive. This page provides the canonical reference for testing SDK classes and methods. For documentation on the client (which is used to manipulate the testnet), see [the Kurtosis Client documentation](../kurtosis-client/lib-documentation).

_Found a bug? File it on [the repo](https://github.com/kurtosis-tech/kurtosis-libs/issues)!_

TestSuiteConfigurator
---------------------
Implementations of this interface are responsible for initializing the user-defined testsuite object for Kurtosis.

### setLogLevel(String logLevelStr)
This function should configure the logging framework that the testsuite uses, since Kurtosis won't know what logging framework the testsuite uses or how to configure it.

**Args**

* `logLevelStr`: The testsuite log level string passed in at runtime, which should be parsed so that the logging framework can be configured.

### parseParamsAndCreateSuite(String paramsJsonStr) -\> [TestSuite][testsuite]
This function should parse the custom testsuite parameters JSON and create an instance of the user's implementation of the `TestSuite` interface.

**Args**

* `paramsJsonStr`: The JSON-serialized custom params data that was passed in when Kurtosis was run, and is used to customize the testsuite's behaviour.

**Returns**

An instance of the user's custom [TestSuite][testsuite] implementation.

Test\<N extends [Network][network]\>
-------------------------------
This interface represents a test that will be executed against a test network. You should create one implementation per test that you want to run. The generic type `N` will be the type of the test network that the test will run against.

### configure([TestConfigurationBuilder][testconfigurationbuilder] builder)
Sets configuration values that will affect the test's execution.

**Args**

* `builder`: The builder that will be used to produce the [TestConfiguration][testconfiguration] defining how the test will behave. You should call the methods on this builder to configure your test.

### setup([NetworkContext][networkcontext] networkContext) -\> N
Performs tasks necessary to initializing the test network before test execution, and returns a [Network][network] implementation that will be fed in as an argument to [Test.run][test_run]. 

For example, to create a test network of three nodes you might call [NetworkContext.addService][networkcontext_addservice] three times here, use [AvailabilityChecker.waitForStartup][availabilitychecker_waitforstartup] to wait for the nodes to be available, and then return the [NetworkContext][networkcontext] (which is a [Network][network] implementation).

For a more complex use case where you've written a custom [Network][network] implementation that encapsulates startup logic in a `setupNetwork` function, you might call something like:

```
MyCustomNetwork customNetwork = new MyCustomNetwork(networkContext);
customNetwork.setupNetwork();
return customNetwork;
```

**Args**

* `networkContext`: The lowest-level representation of the test network being set up. You can modify it by calling methods on this directly, or wrap it in your own custom [Network][network] implementation and use it that way.

**Returns**

A [Network][network] implementation that will be passed to [Test.run][test_run]. If you don't have a custom implementation, you can return the [NetworkContext][networkcontext] (which implements [Network][network]).

### run(N network)
Executes test logic after [Test.setup][test_setup] has completed. For languages that have explicit error return types (e.g. Go), returning an error from this function indicates a failure; for languages that don't (e.g. Java), throwing an exception indicates the same. These will be marked in the individual languages' APIs.

**Args**

* `network`: A [Network][network] implementation representing the test network that the test is executing against.

### getSetupTimeout() -\> Duration
Declares the timeframe in which [Test.setup][test_setup] must complete, to prevent infinite loop bugs from hanging Kurtosis indefinitely.

**Returns**

The time in which test setup must complete.

### getExecutionTimeout() -\> Duration
Declares the timeframe in which [Test.run][test_run] must complete, to prevent infinite loop bugs from hanging Kurtosis indefinitely.

**Returns**

The time in which test execution must complete.

TestConfiguration
-----------------
Object that contains various configuration parameters controlling how a test behaves, which will be configured by the [TestConfigurationBuilder][testconfigurationbuilder] in the [Test.configure][test_configure] method.

### uint32 testSetupTimeoutSeconds
The amount of time a test has to finish the [Test.setup][test_setup] phase. If the phase doesn't complete in the allotted time, an error will be thrown.

### uint32 testRunTimeoutSeconds
The amount of time a test has to finish the [Test.run][test_run] phase. If the phase doesn't complete in the allotted time, an error will be thrown.

### bool isPartitioningEnabled
Setting this to true allows a test to make use of the [NetworkContext.repartitionNetwork][networkcontext_repartitionnetwork] method. This is a configuration flag (rather than enabled by default) because enabling repartitioning requires spinning up extra sidecar Docker containers, and thus an extra load on the system running Kurtosis.

<!-- TODO change key type to FilesArtifactID -->
### Map\<String, String\> filesArtifactUrls
Mapping of a user-defined key -> URL of a gzipped TAR whose contents the test will mount on a service. This should be left empty if no files artifacts are needed. For more details on what files artifacts are, see [ContainerCreationConfig.filesArtifactMountpoints][containercreationconfig_filesartifactmountpoints].

TestConfigurationBuilder
------------------------
Builder for creating a [TestConfiguration][testconfiguration] object, which you should manipulate in your test's [Test.configure][test_configure] function. The functions on this builder will correspond to the properties on the [TestConfiguration][testconfiguration] object, in the form `withProperyName` (e.g. `withSetupTimeoutSeconds` sets the test timeout in seconds). If not set, the default values for the properties are as follows:

* **Test setup timeout seconds:** 180
* **Test run timeout seconds:** 180
* **Partioning enabled:** false
* **Files artifact URLS:** none

TestSuite
---------
Implementations of this interface serve as packages for a set of tests.

### getTests() -\> Map\<String, [Test][test]\>
Returns the tests the testsuite contains. This output can be modified based on custom testsuite parameters (e.g. have a `doSlowTests` flag that can be set to false during local development).

**Returns**

Map of test name -> test object.

### getNetworkWidthBits() -\> uint32
Determines the width (in bits) of the Docker network that Kurtosis will create for each test. The maximum number of IP addresses that any test can use will be 2 ^ network_width_bits, which determines the maximum number of services that can be running at any given time in a testnet. This number should be set high enough that no test will run out of IP addresses, but low enough that the Docker environment doesn't run out of IP addresses (`8` is a good value to start with).

### getStaticFiles() -\> Map\<String, String\>
Defines the static files inside the testsuite container that are available for use when starting services. 

**Returns**

Map of user-defined ID -> absolute filepath inside the testsuite container where the file lives. The user-defined ID is arbitrary, and will be used when starting a service that wants to use the corresponding static file.


---

_Found a bug? File it on [the repo](https://github.com/kurtosis-tech/kurtosis-libs/issues)!_


<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[containerconfigfactory]: ../kurtosis-client/lib-documentation##containerconfigfactorys-extends-service
[containerconfigfactory_getrunconfig]: ../kurtosis-client/lib-documentation##getrunconfigstring-containeripaddr-mapstring-string-generatedfilefilepaths---containerrunconfig

[containercreationconfig]: ../kurtosis-client/lib-documentation#containercreationconfig
[containercreationconfig_usedports]: ../kurtosis-client/lib-documentation#setstring-usedports
[containercreationconfig_filegeneratingfuncs]: ../kurtosis-client/lib-documentation#mapstring-funcfile-filegeneratingfuncs
[containercreationconfig_filesartifactmountpoints]: ../kurtosis-client/lib-documentation#mapstring-string-filesartifactmountpoints
[containercreationconfig_servicecreatingfunc]: ../kurtosis-client/lib-documentation#funcservicecontext---s-servicecreatingfunc

[containercreationconfigbuilder]: ../kurtosis-client/lib-documentation#containercreationconfigbuilder

[containerrunconfig]: ../kurtosis-client/lib-documentation#containerrunconfig

[containerrunconfigbuilder]: ../kurtosis-client/lib-documentation#containerrunconfigbuilder

[network]: ../kurtosis-client/lib-documentation#network

[networkcontext]: ../kurtosis-client/lib-documentation#networkcontext
[networkcontext_addservice]: ../kurtosis-client/lib-documentation#addserviceserviceid-serviceid-containerconfigfactorys-configfactory---s-service-mapstring-portbinding-hostportbindings-availabilitychecker-checker
[networkcontext_addservicetopartition]: ../kurtosis-client/lib-documentation#addservicetopartitionserviceid-serviceid-partitionid-partitionid-containerconfigfactorys-configfactory---s-service-mapstring-portbinding-hostportbindings-availabilitychecker-checker
[networkcontext_repartitionnetwork]: ../kurtosis-client/lib-documentation#repartitionnetworkmappartitionid-setserviceid-partitionservices-mappartitionid-mappartitionid-partitionconnectioninfo-partitionconnections-partitionconnectioninfo-defaultconnection

[partitionconnectioninfo]: ../kurtosis-client/lib-documentation#partitionconnectioninfo

[servicecontext]: ../kurtosis-client/lib-documentation##servicecontext

[generatedfilefilepaths]: ../kurtosis-client/lib-documentation##generatedfilefilepaths

[test]: #testn-extends-network
[test_configure]: #configuretestconfigurationbuilder-builder
[test_setup]: #setupnetworkcontext-networkcontext---n
[test_run]: #runn-network
[test_gettestconfiguration]: #gettestconfiguration---testconfiguration

[testconfiguration]: #testconfiguration

[testconfigurationbuilder]: #testconfigurationbuilder

[testsuite]: #testsuite
