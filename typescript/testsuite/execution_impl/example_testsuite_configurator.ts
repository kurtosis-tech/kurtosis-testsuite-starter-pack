import { ExampleTestsuite } from "../testsuite_impl/example_testsuite";
import { TestSuite } from "kurtosis-testsuite-api-lib";
import * as log from "loglevel";
import { Result, err, ok } from "neverthrow";
import { ExampleTestsuiteArgs } from "./example_testsuite_args";


export class ExampleTestsuiteConfigurator {

    private static safeJsonParse = Result.fromThrowable(JSON.parse, ExampleTestsuiteConfigurator.parseUnknownExceptionValueToError);
    
    constructor () {}
    
    public setLogLevel(logLevelStr: string): Result<null, Error> {
        log.setLevel(<log.LogLevelDesc>logLevelStr);

        return ok(null);
    }

    public parseParamsAndCreateSuite(paramsJsonStr: string): Result<TestSuite, Error> {       
        const argsResult: Result<ExampleTestsuiteArgs, Error> = ExampleTestsuiteConfigurator.safeJsonParse(paramsJsonStr);
        if (argsResult.isErr()) {
            return err(argsResult.error);
        }
        const args: ExampleTestsuiteArgs = argsResult.value;
        
        const validateArgsResult: Result<null, Error> = validateArgs(args);
        if (!validateArgsResult.isOk()) {
            return err(validateArgsResult.error);
        }
        
        const suite: ExampleTestsuite = new ExampleTestsuite(args.apiServiceImage, args.datastoreServiceImage);
        return ok(suite);
    }

    private static parseUnknownExceptionValueToError(value: unknown): Error {
        if (value instanceof Error) {
            return value;
        }
        return new Error("Received an unknown exception value that wasn't an error: " + value);
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