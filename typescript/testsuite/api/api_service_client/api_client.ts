//TODO - extract this to example-microservices repo

import { Result, ok, err } from "neverthrow";
import * as httpStatusCode from "http-status-codes";
import * as axios from "axios";

const PERSON_ENDPOINT: string = "person";

const INCREMENT_BOOKS_READ_ENDPOINT: string = "incrementBooksRead";

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

export interface Person {
    readonly booksRead: number;
}

export class APIClient {
    private readonly ipAddr: string;
    private readonly port: number;
    private static safeJsonParse = Result.fromThrowable(JSON.parse, APIClient.parseUnknownExceptionValueToError);
    private static safeJsonStringify = Result.fromThrowable(JSON.stringify, APIClient.parseUnknownExceptionValueToError);

    constructor (ipAddr: string, port: number) {
        this.ipAddr = ipAddr;
        this.port = port;
    }
    public async addPerson(id: number): Promise <Result<null, Error>> {
        const url: string = this.getPersonUrlForId(id);
        const resp: axios.AxiosResponse<any> = await axios.default.post(url, null);

        return ok(null);
    }

    public async getPerson(id: number): Promise<Result<Person, Error>> {
        const url: string = this.getPersonUrlForId(id);
        let resp: axios.AxiosResponse<any>;
        try {
            resp = await axios.default.get(url);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Performing a get request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        //Note: axios automatically parses the JSON response
        const person: Person = resp.data;
        
        return ok(person);
    }

    public async incrementBooksRead(id: number): Promise<Result<null, Error>> {
        const url: string = "http://" + this.ipAddr + ":" + this.port + "/"+ INCREMENT_BOOKS_READ_ENDPOINT +"/" + id;
        const resp: axios.AxiosResponse<any> = await axios.default.post(url, null);

        return ok(null);
    }

    /*
    Wait for healthy response
    */
    public async waitForHealthy(retries: number, retriesDelayMilliseconds: number): Promise<Result<null, Error>> {

        const url: string = "http://"+this.ipAddr+":"+this.port+"/"+HEALTHCHECK_URL_SLUG;
        let respResult: Result<axios.AxiosResponse<any>, Error> | null = null;

        for (let i = 0 ; i < retries ; i++) {
            respResult = await APIClient.makeHttpGetRequest(url);
            if (respResult.isOk()) {
                break;
            }
            await new Promise(resolve => setTimeout(resolve, retriesDelayMilliseconds));
        }

        if (respResult === null) {
            return err(new Error("Expected a response or error wrapped around Result, but got null instead. Ensure that retries is greater than 0."));
        }
        if (!respResult.isOk()){
            return err(respResult.error);
        }

        const resp: axios.AxiosResponse<any> = respResult.value;
        let bodyString: string = String(resp.data);

        if (bodyString !== HEALTHY_VALUE) {
            return err(new Error("Expected response body text '" + HEALTHY_VALUE + "' from endpoint '" + url + "' but got '" + bodyString +  "' instead"));
        }

        return ok(null);
    }

    private getPersonUrlForId(id: number): string {
        return "http://" + this.ipAddr + ":" + this.port + "/" + PERSON_ENDPOINT + "/" + id;
    }

    private static async makeHttpGetRequest(url: string): Promise<Result<axios.AxiosResponse<any>, Error>>{
        let resp: axios.AxiosResponse<any>;
        try {
            resp = await axios.default.get(url);
        } catch (e: any) {
            return err(e as Error);
        }
        return ok(resp);
    }

    private static parseUnknownExceptionValueToError(value: unknown): Error {
        if (value instanceof Error) {
            return value;
        }
        return new Error("Received an unknown exception value that wasn't an error: " + value);
    }
}