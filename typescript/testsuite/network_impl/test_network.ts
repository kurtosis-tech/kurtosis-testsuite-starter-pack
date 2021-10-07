import { APIClient } from "../api/api_service_client/api_client";
import { DatastoreClient } from "../datastore/datastore_service_client/datastore_client";
import {
    ServiceID,
    NetworkContext,
    ServiceContext,
    PortBinding,
    SharedPath,
    ContainerConfig,
    ContainerConfigBuilder
} from "kurtosis-core-api-lib";
import { Result, ok, err } from "neverthrow";
import * as log from "loglevel";
import {writeFileSync} from "fs";

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

        const datastoreContainerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> = TestNetwork.getDatastoreContainerConfigSupplier()

        const addServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await this.networkCtx.addService(DATASTORE_SERVICE_ID, datastoreContainerConfigSupplier);
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

        const apiServiceContainerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> = TestNetwork.getApiServiceContainerConfigSupplier(this.datastoreClient)
    
        const addServiceResult: Result<[ServiceContext, Map<string, PortBinding>], Error> = await this.networkCtx.addService(serviceId, apiServiceContainerConfigSupplier);
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
