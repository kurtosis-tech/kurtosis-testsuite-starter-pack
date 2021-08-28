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

        let addServiceDatastoreResult: Result<[ServiceContext, Map<string, PortBinding>], Error>;
        try {
            addServiceDatastoreResult = await networkCtx.addService(DATASTORE_SERVICE_ID, containerCreationConfig, runConfigFunc);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling addService method on NetworkContext class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        if (!addServiceDatastoreResult.isOk()) {
            return err(addServiceDatastoreResult.error);
        }
        const [serviceContext, hostPortBindings]: [ServiceContext, Map<string, PortBinding>] = addServiceDatastoreResult.value;

        const datastoreClient: DatastoreClient = new DatastoreClient(serviceContext.getIPAddress(), DATASTORE_PORT);

        let datastoreWaitForHealthyResult: Result<null, Error>;
        try {
            datastoreWaitForHealthyResult = await datastoreClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling waitForHealthy method on DatastoreClient class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        if (!datastoreWaitForHealthyResult.isOk()) {
            return err(datastoreWaitForHealthyResult.error);
        }

        log. info("Added datastore service with host port bindings: ", hostPortBindings)
        return ok(networkCtx);
    }

    public async run(network: Network): Promise<Result<null, Error>> {
        // TODO delete when Test is parameterized with the type of network
        const castedNetwork: NetworkContext = <NetworkContext>network;

        let serviceContextResult: Result<ServiceContext, Error>;
        try {
            serviceContextResult = await castedNetwork.getServiceContext(DATASTORE_SERVICE_ID);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling getServiceContext method on NetworkContext class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        if (!serviceContextResult.isOk()) {
            return err(serviceContextResult.error);
        }
        const serviceContext: ServiceContext = serviceContextResult.value;

        const datastoreClient: DatastoreClient = new DatastoreClient(serviceContext.getIPAddress(), DATASTORE_PORT);

        log.info("Verifying that key '" + TEST_KEY + "' doesn't already exist...");
        let existsResult: Result<boolean, Error>;
        try {
            existsResult = await datastoreClient.exists(TEST_KEY);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling exists method on DatastoreClient class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        if (!existsResult.isOk()) {
            return err(existsResult.error);
        }
        const exists: boolean = existsResult.value;
        if (exists === true) {
            return err(new Error("Test key should not exist yet"));
        }
        log.info("Confirmed that key '" + TEST_KEY + "' doesn't already exist");

        log.info("Inserting value '" + TEST_KEY + "' at key '" + TEST_VALUE + "'...");
        let upsertResult: Result<null, Error>;
        try {
            upsertResult = await datastoreClient.upsert(TEST_KEY, TEST_VALUE);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling upsert method on DatastoreClient class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        if (!upsertResult.isOk()) {
            return err(upsertResult.error);
        }
        log.info("Inserted value successfully");

        log.info("Getting the key we just inserted to verify the value...");
        let getResult: Result<string, Error>;
        try {
            getResult = await datastoreClient.get(TEST_KEY);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Calling get method on DatastoreClient class threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
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