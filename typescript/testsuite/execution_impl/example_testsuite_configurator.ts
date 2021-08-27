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
            // Sadly, we have to do this because there's no great way to enforce the caught thing being an error
            // See: https://stackoverflow.com/questions/30469261/checking-for-typeof-error-in-js
            if (jsonErr && jsonErr.stack && jsonErr.message) {
                return err(jsonErr as Error);
            }
            return err(new Error("Parsing paramsJson string '" + paramsJsonStr + "' threw an exception, but " +
                "it's not an Error so we can't report any more information than this"));

        }
        
        const validateArgsResult: Result<null, Error> = validateArgs(args);
        if (!validateArgsResult.isOk()) {
            return err(validateArgsResult.error);
        }
        
        const suite: ExampleTestsuite = new ExampleTestsuite(args.apiServiceImage, args.datastoreServiceImage);
        return ok(suite);
    }
}

function validateArgs(args: ExampleTestsuiteArgs): Result<null, Error> {
    if (args.apiServiceImage.trim() === "") {
        return err(new Error("API service image is empty"));
    }
    if (args.datastoreServiceImage.trim() === "") {
        return err(new Error("Datastore service image is empty"));
    }
    return ok(null);
}