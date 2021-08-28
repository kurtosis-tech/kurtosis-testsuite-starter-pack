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

    constructor (ipAddr: string, port: number) {
        this.ipAddr = ipAddr;
        this.port = port;
    }
    public async addPerson(id: number): Promise <Result<null, Error>> {
        const url: string = this.getPersonUrlForId(id);
        let resp: axios.AxiosResponse<any>;
        try {
            resp = await axios.default.post(url, null);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Performing a post request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

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

        const body: any = resp.data;
        
        let bodyString: string;
        try {
            bodyString = JSON.stringify(body);
        } catch(jsonErr: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Stringify-ing response data threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        let person: Person;
        try {
            person = JSON.parse(bodyString); 
        } catch(jsonErr: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Parsing body string '" + resp.data + "' threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        
        return ok(person);
    }

    public async incrementBooksRead(id: number): Promise<Result<null, Error>> {
        const url: string = "http://" + this.ipAddr + ":" + this.port + "/"+ INCREMENT_BOOKS_READ_ENDPOINT +"/" + id;
        let resp: axios.AxiosResponse<any>;
        try {
            resp = await axios.default.post(url, null);
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Performing a post request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }

        return ok(null);
    }

    /*
    Wait for healthy response
    */
    public async waitForHealthy(retries: number, retriesDelayMilliseconds: number): Promise<Result<null, Error>> {

        const url: string = "http://"+this.ipAddr+":"+this.port+"/"+HEALTHCHECK_URL_SLUG;
        let respResult: Result<axios.AxiosResponse<any>, Error> | null = null;

        for (let i = 0 ; i < retries ; i++) {
            try {
                respResult = await this.makeHttpGetRequest(url);
            } catch(exception: any) {
                // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
                // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
                if (exception && exception.stack && exception.message) {
                    return err(exception as Error);
                }
                return err(new Error("Making a HTTP get request threw an exception, but " +
                    "it's not an Error so we can't report any more information than this"));
            }
            if (respResult.isOk()) {
                break;
            }
            try {
                await new Promise(resolve => setTimeout(resolve, retriesDelayMilliseconds));
            } catch(exception: any) {
                // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
                // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
                if (exception && exception.stack && exception.message) {
                    return err(exception as Error);
                }
                return err(new Error("Creating a promise for the timeout threw an exception, but " +
                    "it's not an Error so we can't report any more information than this"));
            }
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

    public getPersonUrlForId(id: number): string {
        return "http://" + this.ipAddr + ":" + this.port + "/" + PERSON_ENDPOINT + "/" + id;
    }

    public async makeHttpGetRequest(url: string): Promise<Result<axios.AxiosResponse<any>, Error>>{
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

        return ok(resp);
    }
}