import { Result, err, ok } from "neverthrow";
import * as axios from "axios";
import * as httpStatusCode from "http-status-codes";

const TEXT_CONTENT_TYPE: string = "text/plain"; //TODO (Ali) - not used

const KEY_ENDPOINT: string = "key";

// Use low timeout, so that tests that need timeouts (like network partition tests) will complete quickly
const TIMEOUT_SECONDS: number = 2; //TODO (Ali) - not used

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

class DatastoreClient {
	private readonly ipAddr: string;
	private readonly port: number;

	constructor (ipAddr: string, port: number) {
		this.ipAddr = ipAddr;
		this.port = port;
	}
	
	/*
	Get client's IP address value
	*/
	public getIpAddr(): string { //TODO (Ali) - changed method name, this is okay right?
		return this.ipAddr;
	}

	/*
	Get client's port value
	*/
	public getPort(): number {
		return this.port;
	}
	/*
	Checks if a given key Exists
	*/
	public async exists(key: string): Promise<Result<boolean, Error>> {
		const url: string = this.getUrlForKey(key);
        let resp: axios.AxiosResponse<any>;
        try {
			resp = await axios.default.get(url);
		} catch(exception) {
			return err(exception);
		}
		
		if (resp.status === httpStatusCode.StatusCodes.OK) {
			return ok(true);
		} else if (resp.status === httpStatusCode.StatusCodes.NOT_FOUND) {
			return ok(false);
		} else {
			return err(new Error("Got an unexpected HTTP status code: " + resp.status));
		}
	}
	
	/*
	Gets the value for a given key from the datastore
	*/
	public async get(key: string): Promise<Result<string, Error>> {
		const url: string = this.getUrlForKey(key);
        let resp: axios.AxiosResponse<any>;
        try {
			resp = await axios.default.get(url);
		} catch(exception) {
			return err(exception);
		}
		
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("A non-" + resp.status + " status code was returned"));
		}
		const body: any = resp.data;

        let bodyString: string;
		try {
			bodyString = JSON.stringify(resp.data);
		} catch(jsonErr) {
			return err(jsonErr);
		}

		return ok(bodyString);
	}

	/*
	Puts a value for the given key into the datastore
	*/
	public async upsert(key: string, value: string): Promise<Result<null, Error>> {
		const url: string = this.getUrlForKey(key); 
        let resp: axios.AxiosResponse<any>;
        try {
			resp = await axios.default.post(url, value);
		} catch(exception) {
			return err(exception);
		}

		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("A non-" + resp.status + " status code was returned"));
		}
		return ok(null);
	}

	public getUrlForKey(key: string): string { //TODO (Ali) - since async functions use it, I might need to make this async
		return "http://"+this.ipAddr+":"+this.port+"/"+KEY_ENDPOINT+"/"+key;
	}

	/*
	Wait for healthy response
	*/
	public async waitForHealthy(retries: number, retriesDelayMilliseconds: number): Promise<Result<null, Error>> {

		const url: string = "http://"+this.ipAddr+":"+this.port+"/"+HEALTHCHECK_URL_SLUG;
		let respResult: Result<axios.AxiosResponse<any>, Error>;

		for (let i = 0 ; i < retries ; i++) {
			respResult = await this.makeHttpGetRequest(url);
			if (respResult.isOk()) {
				break;
			}
			await new Promise(f => setTimeout(f, retriesDelayMilliseconds));
		}

		if (!respResult.isOk()){
			return err(respResult.error);
		}

		const resp: axios.AxiosResponse<any> = respResult.value;
        let bodyString: string;
		try {
			bodyString = JSON.stringify(resp.data);
		} catch(jsonErr) {
			return err(jsonErr);
		}

		if (bodyString !== HEALTHY_VALUE) {
			return err(new Error("Expected response body text '" + HEALTHY_VALUE + "' from endpoint '" + url + "' but got '" + bodyString +  "' instead"));
		}

		return ok(null);
	}

	public async makeHttpGetRequest(url: string): Promise<Result<axios.AxiosResponse<any>, Error>>{
		let resp: axios.AxiosResponse<any>;
        try {
			resp = await axios.default.get(url);
		} catch(exception) {
			return err(exception);
		}

		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Received non-OK status code: '" + resp.status + "'"));
		}
		return ok(resp);
	}

}