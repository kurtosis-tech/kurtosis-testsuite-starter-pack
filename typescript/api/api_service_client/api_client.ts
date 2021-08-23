//TODO - extract this to example-microservices repo

//"io/ioutil"
import { Result, ok, err } from "neverthrow";
import * as httpStatusCode from "http-status-codes"
import * as axios from "axios";

const PERSON_ENDPOINT: string = "person";
const TEXT_CONTEXT_TYPE: string = "text/plain"; //TODO (Ali)

const TIMEOUT_SECONDS: number = 2; //TODO (Ali)
const INCREMENT_BOOKS_READ_ENDPOINT: string = "incrementBooksRead";

const HEALTHCHECK_URL_SLUG = "health";
const HEALTHY_VALUE: string = "healthy";

class Person {
	private readonly booksRead: number;

	constructor() {}
}

export class APIClient {
	//httpClient http.Client
	private readonly ipAddr: string;
	private readonly port: number;

    constructor (ipAddr: string, port: number) {
		this.ipAddr = ipAddr;
		this.port = port;
    }

	public async addPerson(id: number): Promise<Result<null, Error>> {
		const url: string = this.getPersonUrlForId(id);
		const resp: axios.AxiosResponse<any> = await axios.default.post(url, null); //TODO (Ali) - might need to catch error to make up for line below ; content type missing in POST request
		
		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err != nil {
		// 	return stacktrace.Propagate(err, "An error occurred making the request to add person with ID '%v'", id)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Adding person with ID '" + id +  "' returned non-OK status code " + resp.status + ""));
		}
		return ok(null);
	}

	public async getPerson(id: number): Promise<Result<Person, Error>> {
		const url: string = this.getPersonUrlForId(id);
		const resp: axios.AxiosResponse<any> = await axios.default.get(url); //TODO (Ali) - might need to catch error to make up for line below
		
		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err != nil {
		// 	return Person{}, stacktrace.Propagate(err, "An error occurred making the request to get person with ID '%v'", id)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Getting person with ID '" + id + "' returned non-OK status code " + resp.status + ""));
		}
		const body: any = resp.data; //TODO (Ali) - building this for JSON.parse

		//TODO (Ali) - how do I deal with a response type of <any>, I can't guarantee on it
		// defer body.Close()
		// bodyBytes, err := ioutil.ReadAll(body)
		// if err != nil {
		// 	return Person{}, stacktrace.Propagate(err, "An error occurred reading the response body")
		// }

		let person: Person;
		try {
			person = JSON.parse(body); 
		} catch(jsonErr) {
			return err(jsonErr);
		}
		
		return ok(person);
	}

	public async incrementBooksRead(id: number): Promise<Result<null, Error>> {
		const url: string = "http://" + this.ipAddr + ":" + this.port + "/"+ INCREMENT_BOOKS_READ_ENDPOINT +"/" + id + ""; //TODO (Ali) - shouldn't we be calling getPersonUrlForId
		const resp: axios.AxiosResponse<any> = await axios.default.post(url, null); //TODO (Ali) - might need to catch error to make up for line below ; content type missing in POST request
		
		//TOOD (Ali) - since I removed http.Client struct, I might remove the following error check
		// if err != nil {
		// 	return stacktrace.Propagate(err, "An error occurred making the request to increment the books read of person with ID '%v'", id)
		// }
		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Incrementing the books read of person with ID '" + id + "' returned non-OK status code " + resp.status + ""));
		}
		return ok(null);
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

	public getPersonUrlForId(id: number): string {
		return "http://" + this.ipAddr + ":" + this.port + "/" + PERSON_ENDPOINT + "/" + id + "";
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