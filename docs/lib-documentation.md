Lib Documentation
=================
Kurtosis libs exist in multiple languages and maintaining in-code comments for each of these is prohibitively expensive. This page exists to provide the canonical reference for Kurtosis lib classes and methods.

TestSuiteConfigurator
---------------------
Implementations of this interface are responsible for initializing the user-defined testsuite object for Kurtosis.

### setLogLevel(String logLevelStr)
This function should configure the logging framework that the testsuite uses, since Kurtosis won't know what logging framework the testsuite uses or how to configure it.

**Args**
* logLevelStr: The testsuite log level string passed in at runtime, which should be parsed so that the logging framework can be configured.

### parseParamsAndCreateSuite(String paramsJsonStr) -> [TestSuite](TODO)
This function should parse the custom testsuite parameters JSON and create an instance of the user's implementation of the `TestSuite` interface.

**Args**
* paramsJsonStr: The JSON-serialized custom params data that was passed in when Kurtosis was run, and is used to customize the testsuite's behaviour.

**Returns**
An instance of the user's custom [TestSuite](TODO) implementation.

Network
-------
This interface provides the option to define a higher level of abstraction for manipulating your test network than is provided by [NetworkContext](TODO), so that test-writing is easier. This commonly looks like wrapping several `NetworkContext` methods into a single one - e.g. if you're running a Cassandra cluster that must bootstrap off three nodes, you might define a `Network` implementation with a `startBootstrappers` method that does the gruntwork so each test doesn't need to add the services manually. Each of your tests will then receive this custom implementation in their [run](TODO) method.

NetworkContext
--------------
This Kurtosis-provided class is the lowest-level representation of a test network, and provides methods for inspecting and manipulating the network. All [Network](TODO) implementations will encapsulate an instance of this class.

### addServiceToPartition(ServiceID serviceId, PartitionID partitionId, [DockerContainerInitializer](TODO)<S> initializer) -> (S service, [AvailabilityChecker](TODO) checker)
Starts a new service in the network with the given service ID, inside the partition with the given ID, using the given initializer.

**Args**
* serviceId: The ID that the new service should have.
* partitionId: The ID of the partition that the new service should be started in. This can be left blank to start the service in the default partition if it exists (i.e. if the network hasn't been repartitioned and the default partition removed).
* initializer: The initializer that provides the logic for starting the container for the new service.

**Returns**
* service: The [Service](TODO) implementation representing the new service, as created by the initializer's [DockerContainerInitializer.getServiceWrappingFunc](TODO) method.
* checker: A class for checking if the returned service is available yet, as defined by [Service.isAvailable](TODO). 

### addService(ServiceID serviceId, [DockerContainerInitializer](TODO)<S> initializer) -> (S, [AvailabilityChecker](TODO))
Convenience wrapper around [addServiceToPartition](TODO), that adds the service to the default partition. Note that if the network has been repartitioned and the default partition doesn't exist anymore, this method will fail.

### <S extends [Service](TODO)> getService(ServiceID serviceId) -> S
Gets the [Service](TODO) interface representing the service with the given ID.

**Args**
* serviceId: The ID of the service in the network to get.

**Returns**
The [Service](TODO) implementation representing the service.

### removeService(ServiceID serviceId, uint64 containerStopTimeoutSeconds)
Stops the container with the given service ID and removes it from the network.

**Args**
* serviceId: The ID of the service to remove.
* containerStopTimeoutSeconds: The number of seconds to wait for the container to gracefully stop before hard-killing it.

### repartitionNetwork(Map<PartitionID, Set<ServiceID>> partitionServices, Map<PartitionID, Map<PartitionID, [PartitionConnectionInfo](TODO)>> partitionConnections, [PartitionConnectionInfo](TODO) defaultConnection)
Repartitions the network so that the connections between services match the specified new state. All services currently in the network must be allocated to a new partition. 

**Args**
* partitionServices: A definition of the new partitions in the network, and the services allocated to each partition. A service can only be allocated to a single partition.
* partitionConnections: Definitions of the connection state between the new partitions. If a connection between two partitions isn't defined in this map, the default connection will be used. Connections are not directional, so an error will be thrown if the same connection is defined twice (e.g. `Map[A][B] = someConnectionInfo`, and `Map[B][A] = otherConnectionInfo`).
* defaultConnection: The network state between two partitions that will be used if the connection isn't defined in the partition connections map.

PartitionConnectionInfo
-----------------------
The `PartitionConnectionInfo` class is a plain old object defining the state between two partitions (e.g. whether network traffic is blocked or not). It is auto-generated from a gRPC API, so exploring it in code is the best way to view its properties.

AvailabilityChecker
-------------------
A class returned by [NetworkContext.addService](TODO) when creating a service, that provides a hook for blocking until the newly-created service is available. This allows for a more-performant workflow of 1) start many services without blocking on their availability and 2) wait for them all to become available.

### waitForStartup(Duration timeBetweenPolls, int maxNumRetries) 
Blocks until the timeout is reached or the checker's corresponding service becomes available (as determined by the [Service.isAvailable](TODO) method).

**Args**
* timeBetweenPolls: The time that the checker should wait before calls to [Service.isAvailable](TODO).
* maxNumRetries: The maximum number of failed calls to [Service.isAvailable](TODO) that the checker will allow before returning an error.

DockerContainerInitializer<S extends [Service](TODO)>
-----------------------------------------------------
Interface that instructs Kurtosis how to create a Docker container that will be represented by an instance of your custom [Service](TODO) implementation. The generic type `S` defines the type of the [Service](TODO) implementation that the initializer will produce.

### getDockerImage() -> String
Provides the name of the Docker image that Kurtosis should use when creating the service's container.

**Returns**
A Docker image specifier (e.g. `my-repo/my-image:some-tag-name`).

### getUsedPorts() -> Set<String>
Provides the ports that the container will be listening on.

**Returns**
Set of ports, in the format `NUM/PROTOCOL` (e.g. `80/tcp`, `9090/udp`, etc.).

### getService([ServiceContext](TODO) serviceContext) -> S
You should fill in this method to create an instance of your custom [Service](TODO) type, usually be wrapping the provided [ServiceContext](TODO) object.

**Args**
* serviceContext: Kurtosis' internal representation of the running container, that your [Service](TODO) implementation can call down to for various purposes.

**Returns**
An instance of your custom [Service](TODO) implementation, that your test will use to interact with the service.

### getFilesToGenerate() -> Set<String>
Declares the files that your service needs generated before your service starts. The file keys here (which can be anything you like) simply tell Kurtosis the number of generated files needed, with the actual file contents generated in [DockerContainerInitializer.initializeGeneratedFiles](TODO).

**Returns**
A set of identifiers for the files to generate that are meaningful to you. E.g. if your service needs a config file and a log file, you might return a set with keys `config` and `log`, whose contents will be generated in [DockerContainerInitializer.initializeGeneratedFiles](TODO).

### initializeGeneratedFiles(Map<String, File> generatedFiles)
Allows the user to fill the contents of files that your service needs before starting. 

**Args**
* generatedFiles: A map whose keys correspond to the output of [DockerContainerInitializer.getFilesToGenerate](TODO) and whose values are filepointers, for you to write whatever file contents you please.

<!-- TODO Change the key type to FilesArtifactID???? -->
### getFilesArtifactMountpoints() -> Map<String, String>
Sometimes a service needs files to be available before it starts, but creating those files via [DockerContainerInitializer.initializeGeneratedFiles](TODO) is slow, difficult, or would require committing a very large artifact to the testsuite's Git repo (e.g. starting a service with a 5 GB Postgres database mounted). To ease this pain, Kurtosis allows you to specify URLs of gzipped TAR files that Kurtosis will download, uncompress, and mount inside your service containers. This function will allow you to specify where those artifacts should get mounted on your service's Docker container.

**Returns**
<!-- TODO when we change the way file artifact IDs are specified, update these docs -->
A map of the file artifact ID -> path on the container where the uncompressed artifact contents should be mounted, with the file artifact IDs corresponding matching the files artifacts declared in the [TestConfiguration](TODO) object returned by [Test.getConfiguration](TODO). E.g. if my test declares an artifact called `5gb-database` that lives at `https://my-site.com/test-artifacts/5gb-database.tgz`, I might return the following map from this function to mount the artifact at the `/database` path inside my container: `{"5gb-database": "/database"}`.

### getTestVolumeMountpoint() -> String
Kurtosis uses a Docker volume to keep track of test state, and needs to mount this volume on every container. Kurtosis can't know what filesystem the service image uses or what paths are safe to mount on though, so this function is how you specify where that volume should be mounted. Your implementation of this method should return a filepath that doesn't already exist where Kurtosis can safely mount the Kurtosis volume.

**Returns**
A filepath where Kurtosis can safely mount the test volume on your container.

### getStartCommandOverrides(Map<String, String> generatedFileFilepaths, String ipAddr) -> (Option<List<String>> entrypointArgs, Option<List<String>> cmdArgs)
You often won't control the Docker images that you'll be using in your testnet, and the `ENTRYPOINT` and `CMD` statements hardcoded in their Dockerfiles might not be suitable for what you need. This function allows you to override these statements to your needs. To use the Dockerfile versions without overriding, leave the option empty.

**Returns**
* entrypointArgs: If set, overrides the `ENTRYPOINT` statement in the image's Dockerfile with the given args.
* cmdArgs: If set, overrides the `CMD` statement in the image's Dockerfile with the given args.

Service
-------
This interface represents a service running in a Docker container inside the test network. Much like the [Network](TODO) interface is a developer-customizable abstraction layer around Kurtosis' [NetworkContext](TODO) representation of the testnet, this interface is a developer-customizable abstraction layer around Kurtosis' [ServiceContext](TODO) representation of a service running in a Docker container. For example, an Elasticsearch service running in a container might be represented by an `ElasticsearchService` class that implements this interface with methods like `updateDocument`, `getDocument` and `deleteDocument`.

### isAvailable() -> bool
Returns a boolean indicating whether the service is available for use. This method is how an [AvailabilityChecker](TODO) determines that a service is available.

**Returns**
True if available, false if not.

ServiceContext
--------------
This Kurtosis-provided class is the lowest-level representation of a service running inside a Docker container. It is your handle for retrieving container information and manipulating the container.

<!-- TODO make the return type be ServiceID???? -->
### getServiceId() -> String
Gets the ID that Kurtosis uses to identify the service.

**Returns**
The service's ID.

### getIpAddress() -> String
Gets the IP address of the Docker container that the service is running inside.

**Returns**
The service's IP address.

### execCommand(List<String> command) -> (int exitCode, List<byte> logs)
Uses [Docker exec](https://docs.docker.com/engine/reference/commandline/exec/) functionality to execute a command inside the service's running Docker container.

**Args**
* command: The args of the command to execute in the container.

**Returns**
* exitCode: The exit code of the command.
* logs: The bytes of the command logs. This isn't a string because Kurtosis can't know what text encoding scheme the container uses.

Test<N extends [Network](TODO)>
-------------------------------
This interface represents a test that will be executed against a test network. You should create one implementation per test that you want to run. The generic type `N` will be the type of the test network that the test will run against.

### getTestConfiguration() -> [TestConfiguration](TODO)
Returns a configuration object that modifies how the test will be executed. Advanced features like network partitioning can be enabled here.

**Returns**
A [TestConfiguration](TODO) object defining how the test should be executed.

### setup([NetworkContext](TODO) networkContext) -> N
Performs tasks necessary to initializing the test network before test execution, and returns a [Network](TODO) implementation that will be fed in as an argument to [Test.run](TODO). 

For example, to create a test network of three nodes you might call [NetworkContext.addService](TODO) three times here, use [AvailabilityChecker.waitForStartup](TODO) to wait for the nodes to be available, and then return the [NetworkContext](TODO) (which is a [Network](TODO) implementation).

For a more complex use case where you've written a custom [Network](TODO) implementation that encapsulates startup logic in a `setupNetwork` function, you might call something like:

```
MyCustomNetwork customNetwork = new MyCustomNetwork(networkContext);
customNetwork.setupNetwork();
return customNetwork;
```

**Args**
* networkContext: The lowest-level representation of the test network being set up. You can modify it by calling methods on this directly, or wrap it in your own custom [Network](TODO) implementation and use it that way.

**Returns**
A [Network](TODO) implementation that will be passed to [Test.run](TODO). If you don't have a custom implementation, you can return the [NetworkContext](TODO) (which implements [Network](TODO)).

<!-- TODO get rid of TestContext -->
### run(N network, [TestContext](TODO) testContext)
Executes test logic after [Test.setup](TODO) has completed.

**Args**
* network: A [Network](TODO) implementation representing the test network that the test is executing against.
* testContext: The test context that the test is running in, which can be used to make assertions and report failures.

### getSetupTimeout() -> Duration
Declares the timeframe in which [Test.setup](TODO) must complete, to prevent infinite loop bugs from hanging Kurtosis indefinitely.

**Returns**
The time in which test setup must complete.

### getExecutionTimeout() -> Duration
Declares the timeframe in which [Test.run](TODO) must complete, to prevent infinite loop bugs from hanging Kurtosis indefinitely.

**Returns**
The time in which test execution must complete.

TestConfiguration
-----------------
Class that defines various configuration parameters that can modify how a test behaves.

### bool isPartitioningEnabled
Setting this to true allows a test to make use of the [NetworkContext.repartitionNetwork](TODO) method. This is a configuration flag (rather than enabled by default) because enabling repartitioning requires spinning up extra sidecar Docker containers, and thus an extra load on the system running Kurtosis.

<!-- TODO change key type to FilesArtifactID -->
### Map<String, String> filesArtifactUrls
Mapping of a user-defined key -> URL of a gzipped TAR whose contents the test will mount on a service. This should be left empty if no files artifacts are needed. For more details on what files artifacts are, see [DockerContainerInitializer.getFilesArtifactMountpoints](TODO).

TestContext
-----------
Context handle you can use for controlling the state of test execution (e.g. making assertions).

### fatal(Error err)
Fails the test with the given error.

**Args**
* err: Error to report, indicating why the test failed.

### assertTrue(bool condition, Error err)
Fails the test if the given condition isn't true.

**Args**
* condition: Predicate that must be true for the test to continue.
* err: Error that will be thrown if the predicate is false.

TestSuite
---------
Implementations of this interface serve as packages for a set of tests.

### getTests() -> Map<String, [Test](TODO)>
Returns the tests the testsuite contains. This output can be modified based on custom testsuite parameters (e.g. have a `doSlowTests` flag that can be set to false during local development).

**Returns**
Map of test name -> test object.

### getNetworkWidthBits() -> uint32
Determines the width (in bits) of the Docker network that Kurtosis will create for each test. The maximum number of IP addresses that any test can use will be 2 ^ network_width_bits, which determines the maximum number of services that can be running at any given time in a testnet. This number should be set high enough that no test will run out of IP addresses, but low enough that the Docker environment doesn't run out of IP addresses (`8` is a good value to start with).
