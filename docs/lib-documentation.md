Lib Documentation
=================
Kurtosis libs exist in multiple languages and maintaining in-code comments for each of these is prohibitively expensive. This page exists to provide the canonical reference for Kurtosis lib classes and methods. Note that any comments specific to a language implementation will be found in the library code comments.

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

Network
-------
This interface provides the option to define a higher level of abstraction for manipulating your test network than is provided by [NetworkContext][networkcontext], so that test-writing is easier. This commonly looks like wrapping several [NetworkContext][networkcontext] methods into a single one - e.g. if you're running a Cassandra cluster that must bootstrap off three nodes, you might define a `CassandraNetwork` implementation with a `startBootstrappers` method that does the gruntwork so each test doesn't need to add the services manually. Each of your tests will then receive this custom implementation in their [Test.run][test_run] method.

NetworkContext
--------------
This Kurtosis-provided class is the lowest-level representation of a test network, and provides methods for inspecting and manipulating the network. All [Network][network] implementations will encapsulate an instance of this class.

### addServiceToPartition(ServiceID serviceId, PartitionID partitionId, [DockerContainerInitializer\<S\>][dockercontainerinitializer] initializer) -\> (S service, [AvailabilityChecker][availabilitychecker] checker)
Starts a new service in the network with the given service ID, inside the partition with the given ID, using the given initializer.

**Args**

* `serviceId`: The ID that the new service should have.
* `partitionId`: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the network hasn't been repartitioned and the default partition removed).
* `initializer`: The initializer that provides the logic for starting the container for the new service.

**Returns**

* `service`: The [Service][service] implementation representing the new service, as created by the initializer's [DockerContainerInitializer.getService][dockercontainerinitializer_getservice] method.
* `checker`: A class for checking if the returned service is available yet, as defined by [Service.isAvailable][service_isavailable]. 

### addService(ServiceID serviceId, [DockerContainerInitializer\<S\>][dockercontainerinitializer] initializer) -\> (S service, [AvailabilityChecker][availabilitychecker] checker)
Convenience wrapper around [NetworkContext.addServiceToPartition][networkcontext_addservicetopartition], that adds the service to the default partition. Note that if the network has been repartitioned and the default partition doesn't exist anymore, this method will fail.

### \<S extends [Service][service]\> getService(ServiceID serviceId) -\> S
Gets the [Service][service] interface representing the service with the given ID.

**Args**

* `serviceId`: The ID of the service in the network to get.

**Returns**

The [Service][service] implementation representing the service.

### removeService(ServiceID serviceId, uint64 containerStopTimeoutSeconds)
Stops the container with the given service ID and removes it from the network.

**Args**

* `serviceId`: The ID of the service to remove.
* `containerStopTimeoutSeconds`: The number of seconds to wait for the container to gracefully stop before hard-killing it.

### repartitionNetwork(Map\<PartitionID, Set\<ServiceID\>> partitionServices, Map\<PartitionID, Map\<PartitionID, [PartitionConnectionInfo][partitionconnectioninfo]\>> partitionConnections, [PartitionConnectionInfo][partitionconnectioninfo] defaultConnection)
Repartitions the network so that the connections between services match the specified new state. All services currently in the network must be allocated to a new partition. 

**Args**

* `partitionServices`: A definition of the new partitions in the network, and the services allocated to each partition. A service can only be allocated to a single partition.
* `partitionConnections`: Definitions of the connection state between the new partitions. If a connection between two partitions isn't defined in this map, the default connection will be used. Connections are not directional, so an error will be thrown if the same connection is defined twice (e.g. `Map[A][B] = someConnectionInfo`, and `Map[B][A] = otherConnectionInfo`).
* `defaultConnection`: The network state between two partitions that will be used if the connection isn't defined in the partition connections map.

PartitionConnectionInfo
-----------------------
This class is a plain old object defining the state between two partitions (e.g. whether network traffic is blocked or not). It is auto-generated from a gRPC API, so exploring it in code is the best way to view its properties.

AvailabilityChecker
-------------------
A class returned by [NetworkContext.addService][networkcontext_addservice] when creating a service, that provides a hook for blocking until the newly-created service is available. This allows for a more-performant workflow of 1) start many services without blocking on their availability and 2) wait for them all to become available.

### waitForStartup(Duration timeBetweenPolls, int maxNumRetries)
Blocks until the timeout is reached or the checker's corresponding service becomes available (as determined by the [Service.isAvailable][service_isavailable] method).

**Args**

* `timeBetweenPolls`: The time that the checker should wait before calls to [Service.isAvailable][service_isavailable].
* `maxNumRetries`: The maximum number of failed calls to [Service.isAvailable][service_isavailable] that the checker will allow before returning an error.

DockerContainerInitializer\<S extends [Service][service]\>
-----------------------------------------------------
Interface that instructs Kurtosis how to create a Docker container that will be represented by an instance of your custom [Service][service] implementation. The generic type `S` defines the type of the [Service][service] implementation that the initializer will produce.

### getDockerImage() -\> String
Provides the name of the Docker image that Kurtosis should use when creating the service's container.

**Returns**

A Docker image specifier (e.g. `my-repo/my-image:some-tag-name`).

### getUsedPorts() -\> Set\<String\>
Provides the ports that the container will be listening on.

**Returns**

Set of ports, in the format `NUM/PROTOCOL` (e.g. `80/tcp`, `9090/udp`, etc.).

### getService([ServiceContext][servicecontext] serviceContext) -\> S
You should fill in this method to create an instance of your custom [Service][service] type, usually be wrapping the provided [ServiceContext][servicecontext] object.

**Args**

* `serviceContext`: Kurtosis' internal representation of the running container, that your [Service][service] implementation can call down to for various purposes.

**Returns**

An instance of your custom [Service][service] implementation, that your test will use to interact with the service.

### getFilesToGenerate() -\> Set\<String\>
Declares the files that your service needs generated before your service starts. The file keys here (which can be anything you like) simply tell Kurtosis the number of generated files needed, with the actual file contents generated in [DockerContainerInitializer.initializeGeneratedFiles][dockercontainerinitializer_initializegeneratedfiles].

**Returns**

A set of identifiers for the files to generate that are meaningful to you. E.g. if your service needs a config file and a log file, you might return a set with keys `config` and `log`, whose contents will be generated in [DockerContainerInitializer.initializeGeneratedFiles][dockercontainerinitializer_initializegeneratedfiles].

### initializeGeneratedFiles(Map\<String, File\> generatedFiles)
Allows the user to fill the contents of files that your service needs before starting. 

**Args**

* `generatedFiles`: A map whose keys correspond to the output of [DockerContainerInitializer.getFilesToGenerate][dockercontainerinitializer_getfilestogenerate] and whose values are filepointers, for you to write whatever file contents you please.

<!-- TODO Change the key type to FilesArtifactID???? -->
### getFilesArtifactMountpoints() -\> Map\<String, String\>
Sometimes a service needs files to be available before it starts, but creating those files via [DockerContainerInitializer.initializeGeneratedFiles][dockercontainerinitializer_initializegeneratedfiles] is slow, difficult, or would require committing a very large artifact to the testsuite's Git repo (e.g. starting a service with a 5 GB Postgres database mounted). To ease this pain, Kurtosis allows you to specify URLs of gzipped TAR files that Kurtosis will download, uncompress, and mount inside your service containers. This function will allow you to specify where those artifacts should get mounted on your service's Docker container.

**Returns**

<!-- TODO when we change the way file artifact IDs are specified, update these docs -->
A map of the file artifact ID -> path on the container where the uncompressed artifact contents should be mounted, with the file artifact IDs corresponding matching the files artifacts declared in the [TestConfiguration][testconfiguration] object returned by [Test.getTestConfiguration][test_gettestconfiguration]. E.g. if my test declares an artifact called `5gb-database` that lives at `https://my-site.com/test-artifacts/5gb-database.tgz`, I might return the following map from this function to mount the artifact at the `/database` path inside my container: `{"5gb-database": "/database"}`.

### getTestVolumeMountpoint() -\> String
Kurtosis uses a Docker volume to keep track of test state, and needs to mount this volume on every container. Kurtosis can't know what filesystem the service image uses or what paths are safe to mount on though, so this function is how you specify where that volume should be mounted. Your implementation of this method should return a filepath that doesn't already exist where Kurtosis can safely mount the Kurtosis volume.

**Returns**

A filepath where Kurtosis can safely mount the test volume on your container.

### getStartCommandOverrides(Map\<String, String\> generatedFileFilepaths, String ipAddr) -\> (Option\<List\<String\>> entrypointArgs, Option\<List\<String\>> cmdArgs)
You often won't control the Docker images that you'll be using in your testnet, and the `ENTRYPOINT` and `CMD` statements hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements to your needs. To use the Dockerfile versions without overriding, leave the option empty.

**Args**

* `generatedFileFilepaths`: Mapping of file ID (as declared in [DockerContainerInitializer.getFilesToGenerate][dockercontainerinitializer_getfilestogenerate]) to the filepath on the container's filesystem where the file exists.
* `ipAddr`: The IP address of the container the service is running inside.

**Returns**

* `entrypointArgs`: If set, overrides the `ENTRYPOINT` statement in the image's Dockerfile with the given args.
* `cmdArgs`: If set, overrides the `CMD` statement in the image's Dockerfile with the given args.

### getEnvironmentVariableOverrides() -\> Map\<String, String\>
Defines environment variables that should be set inside the Docker container running the service. This is often necessary for starting containers from Docker images you don't control, as they'll often be parameterized with environment variables.

**Returns**

A map of environmentVariableName -> environmentVariableValue that will be set inside the container during startup.

Service
-------
This interface represents a service running in a Docker container inside the test network. Much like the [Network][network] interface is a developer-customizable abstraction layer around Kurtosis' [NetworkContext][networkcontext] representation of the testnet, this interface is a developer-customizable abstraction layer around Kurtosis' [ServiceContext][servicecontext] representation of a service running in a Docker container. For example, an Elasticsearch service running in a container might be represented by an `ElasticsearchService` class that implements this interface with methods like `updateDocument`, `getDocument` and `deleteDocument`.

### isAvailable() -\> bool
Returns a boolean indicating whether the service is available for use. This method is how an [AvailabilityChecker][availabilitychecker] determines that a service is available.

**Returns**

True if available, false if not.

ServiceContext
--------------
This Kurtosis-provided class is the lowest-level representation of a service running inside a Docker container. It is your handle for retrieving container information and manipulating the container.

<!-- TODO make the return type be ServiceID???? -->
### getServiceId() -\> String
Gets the ID that Kurtosis uses to identify the service.

**Returns**

The service's ID.

### getIpAddress() -\> String
Gets the IP address of the Docker container that the service is running inside.

**Returns**

The service's IP address.

### execCommand(List\<String\> command) -\> (int exitCode, List\<byte\> logs)
Uses [Docker exec](https://docs.docker.com/engine/reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**

* `command`: The args of the command to execute in the container.

**Returns**

* `exitCode`: The exit code of the command.
* `logs`: The bytes of the command logs. This isn't a string because Kurtosis can't know what text encoding scheme the container uses.

### generateFiles(Set\<String\> filesToGenerate) -\> Map\<String, [GeneratedFileFilepaths][generatedfilefilepaths]\>
Generates files inside the suite execution volume, which is mounted on both the testsuite container and the service container. This allows the testsuite to write data to files that are immediately available to the service container, as if they shared a filesystem.

**Args**

* `filesToGenerate`: A set of user-defined IDs identifying the files that will be generated.

**Returns**

A map of the file IDs (corresponding to the set passed in as input) mapped to a [GeneratedFileFilepaths][generatedfilefilepaths] object containing the filepaths on a) the testsuite container and b) the service container where the generated file was created.

GeneratedFileFilepaths
----------------------
Simple structure containing the filepaths to a generated file on either a) the testsuite container or b) on the service container for whom the file was generated. These filepaths are different because the path where the suite execution volume is mounted on the testsuite container can be different from the path where the volume is mounted on the service container.

### String absoluteFilepathOnTestsuiteContainer
The absolute filepath where the file lives on the testsuite container, which would be used if the testsuite code wants to read or write data to the file.

### String absoluteFilepathOnServiceContainer
The absolute filepath where the file lives on the service container, which would be used if the service wants to read or write data to the file.

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
Executes test logic after [Test.setup][test_setup] has completed. For languages that have explicit error return types (e.g. Go, Rust), returning an error from this function indicates a failure; for languages that don't (e.g. Java), throwing an exception indicates the same. These will be marked in the individual languages' APIs.

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
Mapping of a user-defined key -> URL of a gzipped TAR whose contents the test will mount on a service. This should be left empty if no files artifacts are needed. For more details on what files artifacts are, see [DockerContainerInitializer.getFilesArtifactMountpoints][dockercontainerinitializer_getfilesartifactmountpoints].

TestConfigurationBuilder
------------------------
Builder for creating a [TestConfiguration][testconfiguration] object, which you should manipulate in your test's [Test.configure][test_configure] function. The functions on this builder will correspond to the properties on the [TestConfiguration][testconfiguration] object, in the form `withProperyName` (e.g. `withSetupTimeoutSeconds` sets the test timeout in seconds). If not set, the default values for the properties are as follows:

* **Test setup timeout seconds:** 60
* **Test run timeout seconds:** 60
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


<!-- TODO Make the function definition not include args or return values, so we don't get these huge ugly links that break if we change the function signature -->
<!-- TODO make the reference names a) be properly-cased (e.g. "Service.isAvailable" rather than "service_isavailable") and b) have an underscore in front of them, so they're easy to find-replace without accidentally over-replacing -->

[availabilitychecker]: #availabilitychecker
[availabilitychecker_waitforstartup]: #waitforstartupduration-timebetweenpolls-int-maxnumretries

[dockercontainerinitializer]: #dockercontainerinitializers-extends-service
[dockercontainerinitializer_getfilestogenerate]: #getfilestogenerate---setstring
[dockercontainerinitializer_getservice]: #getserviceservicecontext-servicecontext---s
[dockercontainerinitializer_initializegeneratedfiles]: #initializegeneratedfilesmapstring-file-generatedfiles
[dockercontainerinitializer_getfilesartifactmountpoints]: #getfilesartifactmountpoints---mapstring-string

[network]: #network

[networkcontext]: #networkcontext
[networkcontext_addservice]: #addserviceserviceid-serviceid-dockercontainerinitializers-initializer---s-service-availabilitychecker-checker
[networkcontext_addservicetopartition]: #addservicetopartitionserviceid-serviceid-partitionid-partitionid-dockercontainerinitializers-initializer---s-service-availabilitychecker-checker
[networkcontext_repartitionnetwork]: #repartitionnetworkmappartitionid-setserviceid-partitionservices-mappartitionid-mappartitionid-partitionconnectioninfo-partitionconnections-partitionconnectioninfo-defaultconnection

[partitionconnectioninfo]: #partitionconnectioninfo

[service]: #service
[service_isavailable]: #isavailable---bool

[servicecontext]: #servicecontext

[generatedfilefilepaths]: #generatedfilefilepaths

[test]: #testn-extends-network
[test_configure]: #configuretestconfigurationbuilder-builder
[test_setup]: #setupnetworkcontext-networkcontext---n
[test_run]: #runn-network
[test_gettestconfiguration]: #gettestconfiguration---testconfiguration

[testconfiguration]: #testconfiguration

[testconfigurationbuilder]: #testconfigurationbuilder

[testsuite]: #testsuite
