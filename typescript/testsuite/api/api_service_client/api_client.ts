//TODO - extract this to example-microservices repo

import { Result, ok, err } from "neverthrow";
import * as httpStatusCode from "http-status-codes";
import * as axios from "axios";

const PERSON_ENDPOINT: string = "person";
const TEXT_CONTEXT_TYPE: string = "text/plain"; //TODO (Ali) - not being used

const TIMEOUT_SECONDS: number = 2; //TODO (Ali) - not being used
const INCREMENT_BOOKS_READ_ENDPOINT: string = "incrementBooksRead";

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

export interface Person {
    readonly booksRead: number;

    //TODO remove - if everything works
    // constructor(booksRead: number) {
    //     this.booksRead = booksRead;
    // }

    // public getBooksRead(): number {
    //     return this.booksRead;
    // }
}

export class APIClient {
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
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Performing a post request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
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
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                return err(exception as Error);
            }
            return err(new Error("Performing a get request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
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
        } catch(jsonErr) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Parsing body string '" + bodyString + "' threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        
        return ok(person);
    }

    public async incrementBooksRead(id: number): Promise<Result<null, Error>> {
        const url: string = "http://" + this.ipAddr + ":" + this.port + "/"+ INCREMENT_BOOKS_READ_ENDPOINT +"/" + id;
        let resp: axios.AxiosResponse<any>;
        try {
            resp = await axios.default.post(url, null);
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
        let respResult: Result<axios.AxiosResponse<any>, Error> | null = null;

        for (let i = 0 ; i < retries ; i++) {
            try {
                respResult = await this.makeHttpGetRequest(url);
            } catch(exception) {
                return err(exception);
            }
            if (respResult.isOk()) {
                break;
            }
            try {
                await new Promise(resolve => setTimeout(resolve, retriesDelayMilliseconds));
            } catch(exception) {
                return err(exception);
            }
        }

        if (respResult === null) {
            return err(new Error("Expected a response or error wrapped around Result, but got null instead. Ensure that retries is greater than 0."));
        }
        if (!respResult.isOk()){
            return err(respResult.error);
        }

        const resp: axios.AxiosResponse<any> = respResult.value;

        // let bodyString: string;
        // try {
        //     bodyString = JSON.stringify(resp.data).replace(/['"]+/g, '');
        // } catch(jsonErr) {
        //     // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
        //     // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
        //     if (jsonErr && jsonErr.stack && jsonErr.message) {
        //         return err(jsonErr as Error);
        //     }
        //     return err(new Error("Stringify-ing response data threw an exception, but " +
        //         "it's not an Error so we can't report any more information than this"));
        // }

        if (resp.data !== HEALTHY_VALUE) {
            return err(new Error("Expected response body text '" + HEALTHY_VALUE + "' from endpoint '" + url + "' but got '" + resp.data +  "' instead"));
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