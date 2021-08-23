import { Result, err, ok } from "neverthrow";
//"io/ioutil"
import * as axios from "axios";
import * as httpStatusCode from "http-status-codes";
//"strings"

const TEXT_CONTENT_TYPE: string = "text/plain"; //TODO (Ali)

const KEY_ENDPOINT: string = "key";

// Use low timeout, so that tests that need timeouts (like network partition tests) will complete quickly
const TIMEOUT_SECONDS: number = 2; //TODO (Ali)

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

class DatastoreClient {
	//private readonly httpClient: HttpClient; //TODO (Ali) - might not need since we have axios ;)
	private readonly ipAddr: string;
	private readonly port: number;

	constructor (ipAddr: string, port: number) {
		//this.httpClient = HttpClient(TIMEOUT_SECONDS); //TODO (Ali) - might not need since we have axios ;)
		this.ipAddr = ipAddr;
		this.port = port;
	}
	
	/*
	Get client's IP address value
	*/
	public getIpAddr(): string { //TODO (Ali) - changed method name
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
		const resp: axios.AxiosResponse<any> = await axios.default.get(url); //TODO (Ali) - might need to catch error to make up for line below

		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err !== nil {
		// 	return false, stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
		// }
		
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
		const resp: axios.AxiosResponse<any> = await axios.default.get(url); //TODO (Ali) - might need to do catch error to make up for line below
		
		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err !== nil {
		// 	return "", stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("A non-" + resp.status + " status code was returned"));
		}
		const body: any = resp.data;

		//TODO (Ali) - how do I deal with a response type of <any>, I can't guarantee on it
		// defer body.Close()
		// bodyBytes, err := ioutil.ReadAll(body)
		// if err !== nil {
		// 	return "", stacktrace.Propagate(err, "An error occurred reading the response body")
		// }
		return ok(String(body)); //TODO - we might need to keep this as any type
	}

	/*
	Puts a value for the given key into the datastore
	*/
	public async upsert(key: string, value: string): Promise<Result<null, Error>> {
		const url: string = this.getUrlForKey(key); 
		const resp: axios.AxiosResponse<any> = await axios.default.post(url, value); //TODO (Ali) - might need to catch error to make up for line below; content type missing in POST request

		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// resp, err := client.httpClient.Post(url, textContentType, strings.NewReader(value))
		// if err !== nil {
		// 	return stacktrace.Propagate(err, "An error requesting to upsert data '%v' to key '%v'", value, key)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("A non-" + resp.status + " status code was returned"));
		}
		return ok(null);
	}

	public getUrlForKey(key: string): string { //TODO (Ali) - since async functions use it, I might need to make this async
		return "http://"+this.ipAddr+":"+this.port+"/"+KEY_ENDPOINT+"/"+key+"";
	}

	/*
	Wait for healthy response
	*/
	public async waitForHealthy(retries: number, retriesDelayMilliseconds: number): Promise<Result<null, Error>> {

		const url: string = "http://"+this.ipAddr+":"+this.port+"/"+HEALTHCHECK_URL_SLUG+"";
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

		//TODO (Ali) - how do I deal with a response type of <any>, I can't guarantee on it
		// body := resp.Body
		// defer body.Close()

		// bodyBytes, err := ioutil.ReadAll(body)
		// if err !== nil {
		// 	return stacktrace.Propagate(err, "An error occurred reading the response body")
		// }
		const resp: axios.AxiosResponse<any> = respResult.value;
		const bodyStr: string = String(resp.data); //TODO - we might need to keep this as any type, and remove lines below

		if (bodyStr !== HEALTHY_VALUE) {
			return err(new Error("Expected response body text '" + HEALTHY_VALUE + "' from endpoint '" + url + "' but got '" + bodyStr +  "' instead"));
		}

		return ok(null);
	}

	public async makeHttpGetRequest(url: string): Promise<Result<axios.AxiosResponse<any>, Error>>{
		const resp: axios.AxiosResponse<any> = await axios.default.get(url); //TODO (Ali) - might need to do catch error to make up for line below
		
		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err !== nil {
		// 	return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Received non-OK status code: '" + resp.status + "'"));
		}
		return ok(resp);
	}

}