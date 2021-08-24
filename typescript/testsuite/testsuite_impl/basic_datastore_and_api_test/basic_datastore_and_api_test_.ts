import { ServiceID, NetworkContext, Network, ServiceContext, PortBinding, ContainerRunConfig, StaticFileID, ContainerCreationConfig, ContainerCreationConfigBuilder, ContainerRunConfigBuilder } from "kurtosis-core-api-lib";
import { TestConfigurationBuilder } from "kurtosis-testsuite-api-lib"; //TODO (Ali)
import { Result, err, ok, ResultAsync, errAsync, okAsync } from "neverthrow";
import * as log from "loglevel";
import { DatastoreClient } from "../../../datastore/datastore_service_client/datastore_client";
import * as fs from 'fs';
import { APIClient, Person } from "../../../api/api_service_client/api_client"; //TODO - Extract api client to example-microservices 


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
    
    constructor(datastoreImage: string, apiImage: string) {
        this.datastoreImage = datastoreImage;
        this.apiImage = apiImage;
    }
	
	public configure(builder: TestConfigurationBuilder): void {
        builder.withSetupTimeoutSeconds(60).withRunTimeoutSeconds(60); //TODO (Ali) - allowed since typescript gives direct reference
    }

	public async setup(networkCtx: NetworkContext): Promise<Result<Network, Error>> { //TODO (Ali) - async?

		const [datastoreContainerCreationConfig, datastoreRunConfigFunc]: [ContainerCreationConfig, (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>] = await getDatastoreServiceConfigurations(); //TODO (Ali) - maybe Result here

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

		log.info("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings);

		const [apiServiceContainerCreationConfig, apiServiceRunConfigFunc]: [ContainerCreationConfig, (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>] = await getApiServiceConfigurations(datastoreClient) //TODO (Ali) - add maybe Result here

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

		log.info("Added API service with host port bindings: " + apiSvcHostPortBindings);
		return ok(networkCtx);
	}

	public async run(network: Network): Promise<Result<null, Error>> { //TODO(Ali) - async?
		// Go doesn't have generics so we have to do this cast first
		const castedNetwork: NetworkContext = <NetworkContext>network;

		const serviceContextResult: Result<ServiceContext, Error> = await castedNetwork.getServiceContext(API_SERVICE_ID);
		if (!serviceContextResult.isOk()) {
			return err(serviceContextResult.error);
		}
		const serviceContext: ServiceContext = serviceContextResult.value;

		const apiClient: APIClient = new APIClient(serviceContext.getIPAddress(), API_SERVICE_PORT);

		log.info("Verifying that person with test ID '" + TEST_PERSON_ID + "' doesn't already exist...");
		const getPersonExistsResult: Result<Person, Error> = await apiClient.getPerson(TEST_PERSON_ID);
		if (!getPersonExistsResult.isOk()) {
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

		if (person.getBookRead() !== TEST_NUM_BOOKS_READ) {
			return err(new Error("Expected number of book read '"+TEST_NUM_BOOKS_READ+"' !== actual number of books read '"+person.getBookRead()+"'"));
		}

		return ok(null);
	}
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

//TODO TODO TODO (Ali) - review these helper methods after making final changes to network_impl
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

async function getApiServiceConfigurations(datastoreClient: DatastoreClient): Promise<[ContainerCreationConfig, (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error>]> {
	const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = await getApiServiceConfigInitializingFunc(datastoreClient);

	const apiServiceContainerCreationConfig: ContainerCreationConfig = getApiServiceContainerCreationConfig(configInitializingFunc);

	const apiServiceGenerateRunConfigFunc: (ipAddr: string, generatedFileFilepaths: Map<string, string>, staticFileFilepaths: Map<StaticFileID, string>) => Result<ContainerRunConfig, Error> = await getApiServiceRunConfigFunc();
	return [apiServiceContainerCreationConfig, apiServiceGenerateRunConfigFunc];
}

async function getApiServiceConfigInitializingFunc(datastoreClient: DatastoreClient): Promise<(fp: number) => Promise<Result<null, Error>>> { //Note: Making simplification that file descriptor is just number
	const configInitializingFunc: (fp: number) => Promise<Result<null, Error>> = async (fp: number) => { //TOOD (Ali) - might require changes in ConfigRunFactory in kurt-client due to async
		log.debug("Datastore IP: "+datastoreClient.getIpAddr+" , port: "+datastoreClient.getPort+"");
		const configObj: DatastoreConfig = new DatastoreConfig(datastoreClient.getIpAddr(), datastoreClient.getPort());
		let configBytes: string;
		try { 
			configBytes = JSON.stringify(configObj);
		} catch(jsonErr) {
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