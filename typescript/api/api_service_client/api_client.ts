import { Result, ok, err } from "neverthrow";
import * as httpStatusCode from "http-status-codes";
import * as axios from "axios";

const PERSON_ENDPOINT: string = "person";
const TEXT_CONTEXT_TYPE: string = "text/plain"; //TODO (Ali) - not being used

const TIMEOUT_SECONDS: number = 2; //TODO (Ali) - not being used
const INCREMENT_BOOKS_READ_ENDPOINT: string = "incrementBooksRead";

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

class Person {
	private readonly booksRead: number;

	constructor() {}
}

class APIClient {
	private readonly ipAddr: string;
	private readonly port: number;
	
	constructor (ipAddr: string, port: number) {
		this.ipAddr = ipAddr;
		this.port = port;
	}

	public async addPerson(id: number): Promise<Result<null, Error>> {
		const url: string = this.getPersonUrlForId(id);
		let resp: axios.AxiosResponse<any>;
		try {
			resp = await axios.default.post(url, null);
		} catch(exception) {
			return err(exception);
		}

		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Adding person with ID '" + id +  "' returned non-OK status code " + resp.status));
		}
		return ok(null);
	}

	public async getPerson(id: number): Promise<Result<Person, Error>> {
		const url: string = this.getPersonUrlForId(id);
		let resp: axios.AxiosResponse<any>;
		try {
			resp = await axios.default.get(url);
		} catch(exception) {
			return err(exception);
		}

		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Getting person with ID '" + id + "' returned non-OK status code " + resp.status));
		}
		const body: any = resp.data; //TODO (Ali) - this is a JSON object
        
        //TODO (Ali) - What line below are doing:
        // 1) Taking JSON object and turning it into a string
        // 2) Then parsing the string into a person object 
        // It feels like I'm doing and un-doing an action with these two commands though, but I'm certain I need parse
        // to build the Person object itself from the string
        let bodyString: string;
		try {
			bodyString = JSON.stringify(body);
		} catch(jsonErr) {
			return err(jsonErr);
		}

		let person: Person;
		try {
			person = JSON.parse(bodyString); 
		} catch(jsonErr) {
			return err(jsonErr);
		}
		
		return ok(person);
	}

	public async incrementBooksRead(id: number): Promise<Result<null, Error>> {
		const url: string = "http://" + this.ipAddr + ":" + this.port + "/"+ INCREMENT_BOOKS_READ_ENDPOINT +"/" + id;
		let resp: axios.AxiosResponse<any>;
		try {
			const resp: axios.AxiosResponse<any> = await axios.default.post(url, null);
		} catch(exception) {
			return err(exception);
		}

		if (resp.status !== httpStatusCode.StatusCodes.OK) {
			return err(new Error("Incrementing the books read of person with ID '" + id + "' returned non-OK status code " + resp.status + ""));
		}
		return ok(null);
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
			await new Promise(resolve => setTimeout(resolve, retriesDelayMilliseconds));
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

	public getPersonUrlForId(id: number): string { //TODO (Ali) - since async functions use it, I might need to make this async
		return "http://" + this.ipAddr + ":" + this.port + "/" + PERSON_ENDPOINT + "/" + id;
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