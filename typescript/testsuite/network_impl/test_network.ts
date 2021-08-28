import { APIClient } from "../api/api_service_client/api_client";
import { DatastoreClient } from "../datastore/datastore_service_client/datastore_client";
import { ServiceID, NetworkContext, ContainerCreationConfig, StaticFileID, ContainerRunConfig, ContainerCreationConfigBuilder, ContainerRunConfigBuilder, ServiceContext, PortBinding } from "kurtosis-core-api-lib";
import { Result, ok, err } from "neverthrow";
import * as log from "loglevel";
import * as fs from 'fs';

const DATASTORE_IMAGE: string = "kurtosistech/example-microservices_datastore";
const DATASTORE_SERVICE_ID: ServiceID = "datastore";
const DATASTORE_PORT: number = 1323;

const API_SERVICE_IMAGE: string = "kurtosistech/example-microservices_api";
const API_SERVICE_ID_PREFIX: string = "api-";
const API_SERVICE_PORT: number = 2434;

const WAIT_FOR_STARTUP_DELAY_MILLISECONDS: number = 1000;
const WAIT_FOR_STARTUP_MAX_NUM_POLLS: number = 15;
const CONFIG_FILE_KEY: string = "config-file";

class DatastoreConfig {
    private readonly datastoreIp: string;
    private readonly datastorePort: number;

    constructor(datastoreIp: string, datastorePort: number) {
        this.datastoreIp = datastoreIp;
        this.datastorePort = datastorePort;
    }
}

//  A custom Network implementation is intended to make test-writing easier by wrapping low-level
//  NetworkContext calls with custom higher-level business logic
export class TestNetwork {
    private readonly networkCtx: NetworkContext;
    private readonly datastoreServiceImage: string;
    private readonly apiServiceImage: string;
    private datastoreClient: DatastoreClient | null;
    private personModifyingApiClient: APIClient | null;
    private personRetrievingApiClient: APIClient | null;
    private nextApiServiceId: number;
    private static safeJsonStringify = Result.fromThrowable(JSON.stringify, TestNetwork.parseUnknownExceptionValueToError);

    constructor (networkCtx: NetworkContext, datastoreServiceImage: string, apiServiceImage: string) {
        this.networkCtx = networkCtx;
        this.datastoreServiceImage = datastoreServiceImage;
        this.apiServiceImage = apiServiceImage;
        this.datastoreClient = null;
        this.personModifyingApiClient = null;
        this.personRetrievingApiClient = null;
        this.nextApiServiceId = 0;
    }

    //  Custom network implementations usually have a "setup" method (possibly parameterized) that is used
    //  in the Test.Setup function of each test
    public async setupDatastoreAndTwoApis(): Promise<Result<null, Error>> {

        if (this.datastoreClient !== null) {
            return err(new Error("Cannot add datastore client to network; datastore client already exists!"));
        }

        if (this.personModifyingApiClient !== null || this.personRetrievingApiClient !== null) {
            return err(new Error("Cannot add API services to network; one or more API services already exists"));
        }

        const datastoreContainerCreationConfig: ContainerCreationConfig = TestNetwork.getDataStoreContainerCreationConfig();
        const datastoreRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = TestNetwork.getDataStoreRunConfigFunc();

        const addServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await this.networkCtx.addService(DATASTORE_SERVICE_ID, datastoreContainerCreationConfig, datastoreRunConfigFunc);
        if (!addServiceResult.isOk()) {
            return err(addServiceResult.error);
        }

        const datastoreServiceContext: ServiceContext = addServiceResult.value[0];
        const hostPortBindings: Map<string, PortBinding> = addServiceResult.value[1];

        const datastoreClient: DatastoreClient = new DatastoreClient(datastoreServiceContext.getIPAddress(), DATASTORE_PORT);

        const dataStoreWaitForHealthyResult: Result<null, Error> = await datastoreClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_NUM_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        if (!dataStoreWaitForHealthyResult.isOk()) {
            return err(dataStoreWaitForHealthyResult.error);
        }

        log.info("Added datastore service with host port bindings: ",  hostPortBindings);

        this.datastoreClient = datastoreClient;

        const personModifyingApiClientResult: Result<APIClient, Error> = await this.addApiService();
        if (!personModifyingApiClientResult.isOk()) {
            return err(personModifyingApiClientResult.error);
        }
        this.personModifyingApiClient = personModifyingApiClientResult.value;

        const personRetrievingApiClientResult: Result<APIClient, Error> = await this.addApiService();
        if (!personRetrievingApiClientResult.isOk()) {
            return err(personRetrievingApiClientResult.error);
        }
        this.personRetrievingApiClient = personRetrievingApiClientResult.value;

        return ok(null);
    }

    //  Custom network implementations will also usually have getters, to retrieve information about the
    //  services created during setup
    public getPersonModifyingApiClient(): Result<APIClient, Error> {
        if (this.personModifyingApiClient === null) {
            return err(new Error("No person-modifying API client exists"));
        }
        return ok(this.personModifyingApiClient);
    }

    public getPersonRetrievingApiClient(): Result<APIClient, Error> {
        if (this.personRetrievingApiClient === null) {
            return err(new Error("No person-retrieving API client exists"));
        }
        return ok(this.personRetrievingApiClient)
    }

    public getDatastoreClient(): Result<DatastoreClient, Error>{
        if (this.datastoreClient === null) {
            return err(new Error("No datastore client exists"));
        }
        return ok(this.datastoreClient);
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    
    private async addApiService(): Promise<Result<APIClient, Error>> {

        if (this.datastoreClient === null) {
            return err(new Error("Cannot add API service to network; no datastore client exists"));
        }
    
        const serviceIdStr: string = API_SERVICE_ID_PREFIX + this.nextApiServiceId.toString();
        this.nextApiServiceId = this.nextApiServiceId + 1;
        const serviceId: ServiceID = <ServiceID>(serviceIdStr);
    
        const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = TestNetwork.getApiServiceConfigInitializingFunc(this.datastoreClient);
        const apiServiceContainerCreationConfig: ContainerCreationConfig = TestNetwork.getApiServiceContainerCreationConfig(configInitializingFunc);
        const apiServiceGenerateRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = TestNetwork.getApiServiceRunConfigFunc();
    
        const addServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await this.networkCtx.addService(serviceId, apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc);
        if (!addServiceResult.isOk()) {
            return err(addServiceResult.error);
        }
        const apiServiceContext: ServiceContext = addServiceResult.value[0];
        const hostPortBindings: Map<string, PortBinding> = addServiceResult.value[1];
    
        const apiClient: APIClient = new APIClient(apiServiceContext.getIPAddress(), API_SERVICE_PORT);
    
        const apiClientWaitForHealthyResult: Result<null, Error> = await apiClient.waitForHealthy(WAIT_FOR_STARTUP_MAX_NUM_POLLS, WAIT_FOR_STARTUP_DELAY_MILLISECONDS);
        if (!apiClientWaitForHealthyResult.isOk()) {
            return err(apiClientWaitForHealthyResult.error);
        }
    
        log.info("Added API service with host port bindings:", hostPortBindings)
        return ok(apiClient);
    }

    private static getDataStoreContainerCreationConfig(): ContainerCreationConfig {
        const usedPortsSet: Set<string> = new Set();
        const containerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder(
            DATASTORE_IMAGE,
        ).withUsedPorts(
            usedPortsSet.add(DATASTORE_PORT+"/tcp")
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
            const configBytesResult: Result<string, Error> = TestNetwork.safeJsonStringify(configObj);
            if (configBytesResult.isErr()) {
                return err(configBytesResult.error);
            }
            const configBytes: string = configBytesResult.value;

            log.debug("API config JSON: " + String(configBytes));


            const writeFilePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
                fs.write(fp, configBytes, (error: Error | null) => {
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
