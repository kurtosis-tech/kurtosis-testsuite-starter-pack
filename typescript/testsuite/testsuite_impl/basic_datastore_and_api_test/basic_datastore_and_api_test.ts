import { ServiceID, NetworkContext, Network, ServiceContext, PortBinding, ContainerRunConfig, StaticFileID, ContainerCreationConfig, ContainerCreationConfigBuilder, ContainerRunConfigBuilder } from "kurtosis-core-api-lib";
import { TestConfigurationBuilder } from "kurtosis-testsuite-api-lib";
import { Result, err, ok } from "neverthrow";
import * as log from "loglevel";
import { DatastoreClient } from "../../datastore/datastore_service_client/datastore_client";
import * as fs from 'fs';
import { APIClient, Person } from "../../api/api_service_client/api_client";


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
        
        const datastoreContainerCreationConfig: ContainerCreationConfig = BasicDatastoreAndApiTest.getDataStoreContainerCreationConfig();
        const datastoreRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = BasicDatastoreAndApiTest.getDataStoreRunConfigFunc();

        const addDatastoreServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await networkCtx.addService(DATASTORE_SERVICE_ID, datastoreContainerCreationConfig, datastoreRunConfigFunc);
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

        const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = BasicDatastoreAndApiTest.getApiServiceConfigInitializingFunc(datastoreClient);
        const apiServiceContainerCreationConfig: ContainerCreationConfig = BasicDatastoreAndApiTest.getApiServiceContainerCreationConfig(configInitializingFunc);
        const apiServiceRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = BasicDatastoreAndApiTest.getApiServiceRunConfigFunc();

        const addAPIServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await networkCtx.addService(API_SERVICE_ID, apiServiceContainerCreationConfig, apiServiceRunConfigFunc);
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

    private static getDataStoreContainerCreationConfig(): ContainerCreationConfig {
        const usedPortsSet: Set<string> = new Set();
        const containerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder(
            DATASTORE_IMAGE,
        ).withUsedPorts(
            usedPortsSet.add(DATASTORE_PORT+"/tcp"),
        ).build()
        return containerCreationConfig;
    }

    private static getDataStoreRunConfigFunc(): (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> {
        const runConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = 
        (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => {
            return ok(new ContainerRunConfigBuilder().build());
        }
        return runConfigFunc;
    }

    private static getApiServiceConfigInitializingFunc(datastoreClient: DatastoreClient): (fp: number) => Promise<Result<null, Error>> { //Note: Making simplification that file descriptor is just number
        const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = async (fp: number) => {
            log.debug("Datastore IP: "+datastoreClient.getIpAddr+" , port: "+datastoreClient.getPort);
            const configObj: DatastoreConfig = new DatastoreConfig(datastoreClient.getIpAddr(), datastoreClient.getPort());
            const configBytesResult: Result<string, Error> = BasicDatastoreAndApiTest.safeJsonStringify(configObj);
            if (configBytesResult.isErr()) {
                return err(configBytesResult.error);
            }
            const configBytes: string = configBytesResult.value;

            log.debug("API config JSON: " + String(configBytes));


            const writeFilePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
                fs.writeFile(fp, configBytes, (error: Error | null) => {
                    if (error === null) {
                        resolve(ok(null));
                    } else {
                        resolve(err(error));
                    }
                })
            });
            const writeFileResult: Result<null, Error> = await writeFilePromise;
            if (!writeFileResult.isOk()) {
                return err(writeFileResult.error);
            }
        
            return ok(null);
        }
        return configInitializingFunc;
    }

    private static getApiServiceContainerCreationConfig(configInitializingFunc: (fp: number) => Promise<Result<null, Error>>): ContainerCreationConfig {
        const usedPortsSet: Set<string> = new Set();
        const apiServiceContainerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder(
            API_SERVICE_IMAGE,
        ).withUsedPorts(
            usedPortsSet.add(API_SERVICE_PORT+"/tcp")
        ).withGeneratedFiles(new Map().set(
            CONFIG_FILE_KEY, configInitializingFunc
        )).build();
        return apiServiceContainerCreationConfig;
    }

    private static getApiServiceRunConfigFunc(): (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> {
        const apiServiceRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = 
        (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => {
            if (!generatedFileFilepaths.has(CONFIG_FILE_KEY)) {
                return err(new Error("No filepath found for config file key '"+ CONFIG_FILE_KEY +"'"));
            }
            const configFilepath: string = generatedFileFilepaths.get(CONFIG_FILE_KEY)!;
            const startCmd: string[] = [
                "./api.bin",
                "--config",
                configFilepath
            ]

            const result: ContainerRunConfig = new ContainerRunConfigBuilder().withCmdOverride(startCmd).build();
            return ok(result);
        }
        return apiServiceRunConfigFunc;
    }

    private static parseUnknownExceptionValueToError(value: unknown): Error {
        if (value instanceof Error) {
            return value;
        }
        return new Error("Received an unknown exception value that wasn't an error: " + value);
    }

}