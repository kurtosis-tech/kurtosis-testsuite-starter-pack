//TODO - extract this to example-microservices repo

import { Result, err, ok } from "neverthrow";
import * as axios from "axios";
import * as httpStatusCode from "http-status-codes";

const KEY_ENDPOINT: string = "key";

const HEALTHCHECK_URL_SLUG: string = "health";
const HEALTHY_VALUE: string = "healthy";

export class DatastoreClient {
    private readonly ipAddr: string;
    private readonly port: number;

    constructor (ipAddr: string, port: number) {
        this.ipAddr = ipAddr;
        this.port = port;
    }
    
    /*
    Get client's IP address value
    */
    public getIpAddr(): string { //Note: changed method name from golang
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
        } catch(exception: any) {
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (exception && exception.stack && exception.message) {
                if (axios.default.isAxiosError(exception)) {
                    const exceptionAxiosErr: axios.AxiosError = exception as axios.AxiosError;
                    if (exceptionAxiosErr.response && exceptionAxiosErr.response.status === httpStatusCode.StatusCodes.NOT_FOUND) {
                        return ok(false);
                    }
                }
                return err(exception as Error);
            }
            return err(new Error("Performing a get request threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));
        }
        
        //NOTE: axios will throw an exception on a non-200 status code
        return ok(true);
    }
    
    /*
    Gets the value for a given key from the datastore
    */
    public async get(key: string): Promise<Result<string, Error>> {
        const url: string = this.getUrlForKey(key);
        const resp: axios.AxiosResponse<any> = await axios.default.get(url);
        
        const body: any = resp.data;
        let bodyString: string = String(body);

        return ok(bodyString);
    }

    /*
    Puts a value for the given key into the datastore
    */
    public async upsert(key: string, value: string): Promise<Result<null, Error>> {
        const url: string = this.getUrlForKey(key); 
        const resp: axios.AxiosResponse<any> = await axios.default.post(url, value);

        return ok(null);
    }

    /*
    Wait for healthy response
    */
    public async waitForHealthy(retries: number, retriesDelayMilliseconds: number): Promise<Result<null, Error>> {

        const url: string = "http://"+this.ipAddr+":"+this.port+"/"+HEALTHCHECK_URL_SLUG;
        let respResult: Result<axios.AxiosResponse<any>, Error> | null = null;

        for (let i = 0 ; i < retries ; i++) {
            respResult = await DatastoreClient.makeHttpGetRequest(url);
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

    private getUrlForKey(key: string): string {
        return "http://"+this.ipAddr+":"+this.port+"/"+KEY_ENDPOINT+"/"+key;
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

}