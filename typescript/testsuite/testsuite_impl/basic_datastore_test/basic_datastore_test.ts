import { Network, NetworkContext, ServiceID, ContainerCreationConfig, ContainerCreationConfigBuilder, ContainerRunConfig, StaticFileID, ContainerRunConfigBuilder, ServiceContext, PortBinding } from "kurtosis-core-api-lib";
import { TestConfigurationBuilder } from "kurtosis-testsuite-api-lib";
import * as log from "loglevel";
import { Result, ok, err } from "neverthrow";
import { DatastoreClient } from "../../datastore/datastore_service_client/datastore_client";

const DATASTORE_IMAGE: string = "kurtosistech/example-microservices_datastore";
const DATASTORE_SERVICE_ID: ServiceID = "datastore";
const DATASTORE_PORT: number = 1323;
const TEST_KEY: string = "test-key";
const TEST_VALUE: string = "test-value";

const WAIT_FOR_STARTUP_DELAY_MILLISECONDS: number = 1000;
const WAIT_FOR_STARTUP_MAX_POLLS: number = 15;


export class BasicDatastoreTest {
    private readonly datastoreImage: string;
    
    constructor (datastoreImage: string) {
        this.datastoreImage = datastoreImage;
    }

    public configure(builder: TestConfigurationBuilder) {
        builder.withSetupTimeoutSeconds(60).withRunTimeoutSeconds(60);
    }

    public async setup(networkCtx: NetworkContext): Promise<Result<Network, Error>> {

        const containerCreationConfig: ContainerCreationConfig = BasicDatastoreTest.getContainerCreationConfig();
        const runConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = BasicDatastoreTest.getRunConfigFunc();

        const addServiceDatastoreResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await networkCtx.addService(DATASTORE_SERVICE_ID, containerCreationConfig, runConfigFunc);
        if (!addServiceDatastoreResult.isOk()) {
            return err(addServiceDatastoreResult.error);
        }
        const [serviceContext, hostPortBindings]: [ServiceContext, Map<string, PortBinding>] = addServiceDatastoreResult.value;

        const datastoreClient: DatastoreClient = new DatastoreClient(serviceContext.getIPAddress(), DATASTORE_PORT);

        const datastoreWaitForHealthyResult: Result<null, Error> = await datastoreClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        if (!datastoreWaitForHealthyResult.isOk()) {
            return err(datastoreWaitForHealthyResult.error);
        }

        log. info("Added datastore service with host port bindings: ", hostPortBindings)
        return ok(networkCtx);
    }

    public async run(network: Network): Promise<Result<null, Error>> {
        // TODO delete when Test is parameterized with the type of network
        const castedNetwork: NetworkContext = <NetworkContext>network;

        const serviceContextResult: Result<ServiceContext, Error> = await castedNetwork.getServiceContext(DATASTORE_SERVICE_ID);
        if (!serviceContextResult.isOk()) {
            return err(serviceContextResult.error);
        }
        const serviceContext: ServiceContext = serviceContextResult.value;

        const datastoreClient: DatastoreClient = new DatastoreClient(serviceContext.getIPAddress(), DATASTORE_PORT);

        log.info("Verifying that key '" + TEST_KEY + "' doesn't already exist...");
        const existsResult: Result<boolean, Error> = await datastoreClient.exists(TEST_KEY);
        if (!existsResult.isOk()) {
            return err(existsResult.error);
        }
        const exists: boolean = existsResult.value;
        if (exists === true) {
            return err(new Error("Test key should not exist yet"));
        }
        log.info("Confirmed that key '" + TEST_KEY + "' doesn't already exist");

        log.info("Inserting value '" + TEST_KEY + "' at key '" + TEST_VALUE + "'...");
        const upsertResult: Result<null, Error> = await datastoreClient.upsert(TEST_KEY, TEST_VALUE);
        if (!upsertResult.isOk()) {
            return err(upsertResult.error);
        }
        log.info("Inserted value successfully");

        log.info("Getting the key we just inserted to verify the value...");
        const getResult: Result<string, Error> = await datastoreClient.get(TEST_KEY);
        if (!getResult.isOk()) {
            return err(getResult.error);
        }
        const value: string = getResult.value;
        if (value !== TEST_VALUE) {
            return err(new Error("Returned value '" + value + "' !== test value '" + TEST_VALUE + "'"));
        }
        log.info("Value verified");
        return ok(null);
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================

    private static getContainerCreationConfig(): ContainerCreationConfig {
        const usedPortsSet: Set<string> = new Set();
        const containerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder(
            DATASTORE_IMAGE,
        ).withUsedPorts(
            usedPortsSet.add(DATASTORE_PORT+"/tcp")
        ).build()
        return containerCreationConfig;
    }

    private static getRunConfigFunc(): (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> {
        const runConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = 
        (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => {
            return ok(new ContainerRunConfigBuilder().build());
        }
        return runConfigFunc;
    }
}