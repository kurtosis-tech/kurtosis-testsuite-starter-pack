import { APIClient } from "../../api/api_service_client/api_client";
import { DatastoreClient } from "../../datastore/datastore_service_client/datastore_client"; //TODO - extract to microservice-examples
import { ServiceID, NetworkContext, ContainerCreationConfig, StaticFileID, ContainerRunConfig, ContainerCreationConfigBuilder, ContainerRunConfigBuilder, ServiceContext, PortBinding } from "kurtosis-core-api-lib"; //TODO (Ali)
import { Result, ok, err, ResultAsync, okAsync, errAsync } from "neverthrow";
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
//    NetworkContext calls with custom higher-level business logic
class TestNetwork {
	private readonly networkCtx: NetworkContext;
	private readonly datastoreServiceImage: string;
	private readonly apiServiceImage: string;
	private datastoreClient: DatastoreClient;
	private personModifyingApiClient: APIClient;
	private personRetrievingApiClient: APIClient;
	private nextApiServiceId: number;

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
	//   in the Test.Setup function of each test
	public async setupDatastoreAndTwoApis(): Promise<Result<null, Error>> {

		if (this.datastoreClient != null) {
			return err(new Error("Cannot add datastore client to network; datastore client already exists!"));
		}

		if (this.personModifyingApiClient != null || this.personRetrievingApiClient != null) {
			return err(new Error("Cannot add API services to network; one or more API services already exists"));
		}

		const [datastoreContainerCreationConfig, datastoreRunConfigFunc] = await getDatastoreServiceConfigurations();

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

		log.info("Added datastore service with host port bindings: " + hostPortBindings);

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
	//   services created during setup
	public getPersonModifyingApiClient(): Result<APIClient, Error> {
		if (this.personModifyingApiClient == null) {
			return err(new Error("No person-modifying API client exists"));
		}
		return ok(this.personModifyingApiClient);
	}

	public getPersonRetrievingApiClient(): Result<APIClient, Error> {
		if (this.personRetrievingApiClient == null) {
			return err(new Error("No person-retrieving API client exists"));
		}
		return ok(this.personRetrievingApiClient)
	}

	public getDatastoreClient(): Result<DatastoreClient, Error>{ //TODO (Ali) (comment) - added this getter
		if (this.datastoreClient == null) {
			return err(new Error("No datastore client exists"));
		}
		return ok(this.datastoreClient);
	}

	//Private helper function
	public async addApiService(): Promise<Result<APIClient, Error>> { //TODO (Ali) - All methods inside would need to be turned into promises here

		if (this.datastoreClient == null) {
			return err(new Error("Cannot add API service to network; no datastore client exists"));
		}
	
		const serviceIdStr: string = API_SERVICE_ID_PREFIX + this.nextApiServiceId.toString();
		this.nextApiServiceId = this.nextApiServiceId + 1;
		const serviceId: ServiceID = <ServiceID>(serviceIdStr);
	
		const [apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc] = await getApiServiceConfigurations(this);
	
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
	
		log.info("Added API service with host port bindings:" + hostPortBindings)
		return ok(apiClient);
	}

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
// func (network *TestNetwork)  addApiService() (*api_service_client.APIClient, error) {

// 	if network.datastoreClient == nil {
// 		return nil, stacktrace.NewError("Cannot add API service to network; no datastore client exists")
// 	}

// 	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
// 	network.nextApiServiceId = network.nextApiServiceId + 1
// 	serviceId := services.ServiceID(serviceIdStr)

// 	apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc := getApiServiceConfigurations(network)

// 	apiServiceContext, hostPortBindings, err := network.networkCtx.AddService(serviceId, apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc)
// 	if err != nil {
// 		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
// 	}

// 	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), apiServicePort)

// 	err = apiClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
// 	if err != nil {
// 		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
// 	}

// 	log.info("Added API service with host port bindings:" + hostPortBindings)
// 	return apiClient, nil
// }

async function getDatastoreServiceConfigurations(): Promise<[ContainerCreationConfig, (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>]> {
	const datastoreContainerCreationConfig: ContainerCreationConfig = await getDataStoreContainerCreationConfig();

	const datastoreRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = await getDataStoreRunConfigFunc();
	return [datastoreContainerCreationConfig, datastoreRunConfigFunc];
}

async function getDataStoreContainerCreationConfig(): Promise<ContainerCreationConfig> {
	const containerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder( //TODO (Ali) - may need to make ContainerCreationConfig async
		DATASTORE_IMAGE,
	).withUsedPorts(
		new Set(""+DATASTORE_PORT+"/tcp"),
	).build()
	return containerCreationConfig;
}

async function getDataStoreRunConfigFunc(): Promise<(ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>> {
	const runConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = 
	(ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => {
		return ok(new ContainerRunConfigBuilder().build());
	}
	return runConfigFunc;
}

async function getApiServiceConfigurations(network: TestNetwork): Promise<[ContainerCreationConfig, (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>]> {
	const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = await getApiServiceConfigInitializingFunc(network.getDatastoreClient());

	const apiServiceContainerCreationConfig: ContainerCreationConfig = getApiServiceContainerCreationConfig(configInitializingFunc);

	const apiServiceGenerateRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = await getApiServiceRunConfigFunc();
	return [apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc];
}

async function getApiServiceConfigInitializingFunc(datastoreClientResult: Result<DatastoreClient, Error>): Promise<(fp: number) => Promise<Result<null, Error>>> { //Making simplification that file descriptor is just number
	const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = async (fp: number) => { //TOOD (Ali) - might require changes in ConfigRunFactory in kurt-client
		if (!datastoreClientResult.isOk()) {
			return err(datastoreClientResult.error);
		}
		const datastoreClient: DatastoreClient = datastoreClientResult.value;

		log.debug("Datastore IP: "+datastoreClient.getIpAddr+" , port: "+datastoreClient.getPort+"");
		const configObj: DatastoreConfig = new DatastoreConfig(datastoreClient.getIpAddr(), datastoreClient.getPort());
		let configBytes: string;
		try { 
			configBytes = JSON.stringify(configObj);
		}
		catch(jsonErr) {
			return err(jsonErr);
		}

		log.debug("API config JSON: " + String(configBytes));


		const writeFilePromise: Promise<ResultAsync<null, Error>> = new Promise((resolve, _unusedReject) => {
			fs.writeFile(fp, configBytes, (error: Error) => {
				if (error === null) {
					resolve(okAsync(null));
				} else {
					resolve(errAsync(error));
				}
			})
		});
		const writeFileResult: Result<fs.Stats, Error> = await writeFilePromise;
		if (!writeFileResult.isOk()) {
			return err(writeFileResult.error);
		}
	
		return ok(null);
	}
	return configInitializingFunc;
}

async function getApiServiceContainerCreationConfig(configInitializingFunc: (fp: number) => Promise<Result<null, Error>>): Promise<ContainerCreationConfig> {
	const apiServiceContainerCreationConfig: ContainerCreationConfig = new ContainerCreationConfigBuilder(
		API_SERVICE_IMAGE,
	).withUsedPorts(
		new Set(API_SERVICE_PORT+"/tcp")
	).withGeneratedFiles(new Map().set(
		CONFIG_FILE_KEY, configInitializingFunc //TODO (Ali) - might need to wrap value in kurt client of this func inside a promise
	)).build();
	return apiServiceContainerCreationConfig;
}

async function getApiServiceRunConfigFunc(): Promise<(ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>> {
	const apiServiceRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = 
	(ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => {
		if (!generatedFileFilepaths.has(CONFIG_FILE_KEY)) {
			return err(new Error("No filepath found for config file key '"+ CONFIG_FILE_KEY +"'"));
		}
		const configFilepath: string = generatedFileFilepaths[CONFIG_FILE_KEY];
		const startCmd: string[] = [
			"./api.bin",
			"--config",
			configFilepath
		]

		const result: ContainerCreationConfig = new ContainerRunConfigBuilder().withCmdOverride(startCmd).build();
		return result;
	}
	return apiServiceRunConfigFunc;
}
