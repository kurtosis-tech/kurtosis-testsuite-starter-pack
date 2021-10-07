import {
    ServiceID,
    NetworkContext,
    Network,
    ServiceContext,
    PortBinding,
    SharedPath,
    ContainerConfig,
    ContainerConfigBuilder
} from "kurtosis-core-api-lib";
import { TestConfigurationBuilder } from "kurtosis-testsuite-api-lib";
import { Result, err, ok } from "neverthrow";
import * as log from "loglevel";
import { DatastoreClient } from "../../datastore/datastore_service_client/datastore_client";
import { APIClient, Person } from "../../api/api_service_client/api_client";
import {writeFileSync} from "fs";


const DATASTORE_IMAGE: string = "kurtosistech/example-microservices_datastore";
const DATASTORE_SERVICE_ID: ServiceID = "datastore";
const DATASTORE_PORT: number = 1323;

const API_SERVICE_IMAGE: string = "kurtosistech/example-microservices_api";
const API_SERVICE_ID: ServiceID = "api";
const API_SERVICE_PORT: number = 2434;

const WAIT_FOR_STARTUP_DELAY_MILLISECONDS: number = 1000;
const WAIT_FOR_STARTUP_MAX_POLLS: number = 15;

const TEST_PERSON_ID: number = 23;
const TEST_NUM_BOOKS_READ: number = 3;

const CONFIG_FILE_KEY: string  = "config-file";


class DatastoreConfig {
    private readonly datastoreIp: string;
    private readonly datastorePort: number;
    
    constructor(datastoreIp: string, datastorePort: number) {
        this.datastoreIp = datastoreIp;
        this.datastorePort = datastorePort;
    }
}

export class BasicDatastoreAndApiTest {
    private readonly datastoreImage: string;
    private readonly apiImage: string;
    private static safeJsonStringify = Result.fromThrowable(JSON.stringify, BasicDatastoreAndApiTest.parseUnknownExceptionValueToError);
    
    constructor(datastoreImage: string, apiImage: string) {
        this.datastoreImage = datastoreImage;
        this.apiImage = apiImage;
    }
    
    public configure(builder: TestConfigurationBuilder): void {
        builder.withSetupTimeoutSeconds(60).withRunTimeoutSeconds(60);
    }

    public async setup(networkCtx: NetworkContext): Promise<Result<Network, Error>> {

        const datastoreContainerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> = BasicDatastoreAndApiTest.getDatastoreContainerConfigSupplier()

        const addDatastoreServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await networkCtx.addService(DATASTORE_SERVICE_ID, datastoreContainerConfigSupplier);
        if (!addDatastoreServiceResult.isOk()) {
            return err(addDatastoreServiceResult.error)
        }
        const [datastoreServiceContext, datastoreSvcHostPortBindings]: [ServiceContext, Map<string, PortBinding>] = addDatastoreServiceResult.value;

        const datastoreClient: DatastoreClient = new DatastoreClient(datastoreServiceContext.getIPAddress(), DATASTORE_PORT);

        const datastoreWaitForHealthyResult: Result<null, Error> = await datastoreClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        if (!datastoreWaitForHealthyResult.isOk()) {
            return err(datastoreWaitForHealthyResult.error);
        }

        log.info("Added datastore service with host port bindings: ", datastoreSvcHostPortBindings);

        const apiServiceContainerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> = BasicDatastoreAndApiTest.getApiServiceContainerConfigSupplier(datastoreClient)

        const addAPIServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await networkCtx.addService(API_SERVICE_ID, apiServiceContainerConfigSupplier);
        if (!addAPIServiceResult.isOk()) {
            return err(addAPIServiceResult.error)
        }
        const [apiServiceContext, apiSvcHostPortBindings]: [ServiceContext, Map<string, PortBinding>] = addAPIServiceResult.value;

        const apiClient: APIClient = new APIClient(apiServiceContext.getIPAddress(), API_SERVICE_PORT);

        const apiWaitForHealthyResult: Result<null, Error> = await apiClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        if (!apiWaitForHealthyResult.isOk()) {
            return err(apiWaitForHealthyResult.error);
        }

        log.info("Added API service with host port bindings: ", apiSvcHostPortBindings);
        return ok(networkCtx);
    }

    public async run(network: Network): Promise<Result<null, Error>> {
        // TODO delete when Test is parameterized with the type of network
        const castedNetwork: NetworkContext = <NetworkContext>network;

        const getServiceContextResult: Result<ServiceContext, Error> = await castedNetwork.getServiceContext(API_SERVICE_ID);
        if (!getServiceContextResult.isOk()) {
            return err(getServiceContextResult.error);
        }
        const serviceContext: ServiceContext = getServiceContextResult.value;

        const apiClient: APIClient = new APIClient(serviceContext.getIPAddress(), API_SERVICE_PORT);

        log.info("Verifying that person with test ID '" + TEST_PERSON_ID + "' doesn't already exist...");
        let getPersonExistsResult: Result<Person, Error>
        getPersonExistsResult = await apiClient.getPerson(TEST_PERSON_ID);
        if (getPersonExistsResult.isOk()) {
            return err(new Error("Expected an error trying to get a person who doesn't exist yet, but didn't receive one"));
        }
        log.info("Verified that test person doesn't already exist");

        log.info("Adding test person with ID '" + TEST_PERSON_ID + "'...");
        const addPersonResult: Result<null, Error> = await apiClient.addPerson(TEST_PERSON_ID);
        if (!addPersonResult.isOk()) {
            return err(addPersonResult.error);
        }
        log.info("Test person added");

        log.info("Incrementing test person's number of books read by " + TEST_NUM_BOOKS_READ + "...");
        for (let i = 0; i < TEST_NUM_BOOKS_READ; i++) {
            const incrementBooksReadResult: Result<null, Error> = await apiClient.incrementBooksRead(TEST_PERSON_ID);
            if (!incrementBooksReadResult.isOk()) {
                return err(incrementBooksReadResult.error);
            }
        }
        log.info("Incremented number of books read");

        log.info("Retrieving test person to verify number of books read...");
        const getPersonResult: Result<Person, Error> = await apiClient.getPerson(TEST_PERSON_ID);
        if (!getPersonResult.isOk()) {
            return err(getPersonResult.error);
        }
        const person: Person = getPersonResult.value;
        log.info("Retrieved test person");

        if (person.booksRead !== TEST_NUM_BOOKS_READ) {
            return err(new Error("Expected number of book read '"+TEST_NUM_BOOKS_READ+"' !== actual number of books read '"+person.booksRead+"'"));
        }

        return ok(null);
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================

    private static getDatastoreContainerConfigSupplier(): (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {
        const containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> =
            (ipAddr: string, sharedDirectory: SharedPath) => {
                const usedPortsSet: Set<string> = new Set();
                const containerConfig: ContainerConfig = new ContainerConfigBuilder(
                    DATASTORE_IMAGE
                ).withUsedPorts(
                    usedPortsSet.add(DATASTORE_PORT + "/tcp")
                ).build()
                return ok(containerConfig)
            }
        return containerConfigSupplier
    }

    private static getApiServiceContainerConfigSupplier(datastoreClient: DatastoreClient): (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {
        const containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> =
            (ipAddr: string, sharedDirectory: SharedPath) => {
                const usedPortsSet: Set<string> = new Set();

                const datastoreConfigFileFilePathResult: Result<SharedPath, Error> = this.createDatastoreConfigFileInServiceDirectory(datastoreClient, sharedDirectory);
                if (!datastoreConfigFileFilePathResult.isOk()) {
                    return err(datastoreConfigFileFilePathResult.error);
                }

                const datastoreConfigFileFilePath: SharedPath = datastoreConfigFileFilePathResult.value

                const startCmd: string[] = [
                    "./api.bin",
                    "--config",
                    datastoreConfigFileFilePath.getAbsPathOnServiceContainer()
                ]

                const containerConfig: ContainerConfig = new ContainerConfigBuilder(
                    API_SERVICE_IMAGE
                ).withUsedPorts(
                    usedPortsSet.add(API_SERVICE_PORT + "/tcp")
                ).withCmdOverride(
                    startCmd
                ).build()
                return ok(containerConfig)
            }
        return containerConfigSupplier
    }

    private static createDatastoreConfigFileInServiceDirectory(datastoreClient: DatastoreClient, sharedDirectory: SharedPath): Result<SharedPath, Error> {
        const configFileFilePath: SharedPath = sharedDirectory.getChildPath(CONFIG_FILE_KEY)

        log.info("Config file absolute path on this container: " + configFileFilePath.getAbsPathOnThisContainer() + " , on service container: " + configFileFilePath.getAbsPathOnServiceContainer());

        log.debug("Datastore IP: " + datastoreClient.getIpAddr() + " , port: " + datastoreClient.getPort());

        const configObj: DatastoreConfig = new DatastoreConfig(datastoreClient.getIpAddr(), datastoreClient.getPort())

        let configBytes: string;
        try {
            configBytes = JSON.stringify(configObj);
        } catch (jsonErr: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Stringify-ing DatastoreConfig object threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        log.debug("API config JSON: " + configBytes)

        try {
            writeFileSync(configFileFilePath.getAbsPathOnThisContainer(), configBytes)
        } catch (exception) {
            if (exception instanceof Error) {
                return err(new Error("An error occurred writing the serialized config JSON to file with error: " + exception));
            } else {
                return err(new Error("An unknown exception value was thrown during writing the serialized config JSON to file with error: " + exception))
            }
        }
        return ok(configFileFilePath);
    }

    private static parseUnknownExceptionValueToError(value: unknown): Error {
        if (value instanceof Error) {
            return value;
        }
        return new Error("Received an unknown exception value that wasn't an error: " + value);
    }

}