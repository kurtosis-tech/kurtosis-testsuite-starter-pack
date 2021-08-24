import { ExampleTestsuite } from "../testsuite_impl/example_testsuite";
import { TestSuite } from "kurtosis-testsuite-api-lib"; //TODO (Ali)
import * as log from "loglevel";
import { Result, err, ok } from "neverthrow";
import { ExampleTestsuiteArgs } from "./example_testsuite_args";


class ExampleTestsuiteConfigurator {
    
    construcor () {}
	
	public setLogLevel(logLevelStr: string): Result<null, Error> {
		const newLog: log.Logger = log.getLogger(logLevelStr);
        newLog.setLevel(newLog.getLevel());
        // log.SetFormatter(&logrus.TextFormatter{ //TOOD (Ali) - loglevel doesn't support formatting since it would stop stacktrace information
        //     ForceColors:   true,
        //     FullTimestamp: true,
        // })
        return ok(null);
    }

    public parseParamsAndCreateSuite(paramsJsonStr: string): Result<TestSuite, Error> {
        let args: ExampleTestsuiteArgs;
		try {
			args = JSON.parse(paramsJsonStr); //TODO (Ali) - golang used bytes in their unMarshal
		} catch (jsonErr) {
			return err(jsonErr);
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
	if (args.apiServiceImage.trim() == "") {
		return err(new Error("API service image is empty"));
	}
	if (args.datastoreServiceImage.trim() == "") {
		return err(new Error("Datastore service image is empty"));
	}
	return ok(null);
}