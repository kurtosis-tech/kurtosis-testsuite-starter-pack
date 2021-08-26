import { ExampleTestsuite } from "../testsuite_impl/example_testsuite";
import { TestSuite } from "kurtosis-testsuite-api-lib";
import * as log from "loglevel";
import { Result, err, ok } from "neverthrow";
import { ExampleTestsuiteArgs } from "./example_testsuite_args";


export class ExampleTestsuiteConfigurator {
    
    construcor () {}
    
    public setLogLevel(logLevelStr: string): Result<null, Error> {
        const newLog: log.Logger = log.getLogger(logLevelStr);
        log.setLevel(newLog.getLevel());

        return ok(null);
    }

    public parseParamsAndCreateSuite(paramsJsonStr: string): Result<TestSuite, Error> {
        let args: ExampleTestsuiteArgs;
        try {
            args = JSON.parse(paramsJsonStr);
        } catch (jsonErr) {
            return err(jsonErr);
        }
        
        const validateArgsResult: Result<null, Error> = validateArgs(args);
        if (!validateArgsResult.isOk()) {
            return err(validateArgsResult.error);
        }
        
        const suite: ExampleTestsuite = new ExampleTestsuite(args.getApiServiceImage(), args.getDatastoreServiceImage());
        return ok(suite);
    }
}

function validateArgs(args: ExampleTestsuiteArgs): Result<null, Error> {
    if (args.getApiServiceImage().trim() === "") {
        return err(new Error("API service image is empty"));
    }
    if (args.getDatastoreServiceImage().trim() === "") {
        return err(new Error("Datastore service image is empty"));
    }
    return ok(null);
}