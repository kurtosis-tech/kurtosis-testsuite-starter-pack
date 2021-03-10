Lib Documentation
=================


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
A class









TestSuiteExecutor
-----------------
Kurtosis-provided class responsible for executing a testsuite. You won't need to modify this class - only call `run`.

### new(String kurtosisApiSocket, String logLevelStr, String paramsJsonStr, [TestSuiteConfigurator](TODO) configurator)
Creates a new instance 

### run()
* 


