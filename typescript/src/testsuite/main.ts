import { ExampleTestsuiteConfigurator } from "./testsuite_impl/execution_impl/example_testsuite_configurator";
import { TestSuiteExecutor } from "kurtosis-testsuite-api-lib";
import * as log from "loglevel";
import { Result, err, ok } from "neverthrow";

const SUCCESS_EXIT_CODE: number = 0;
const FAILURE_EXIT_CODE: number = 1;

async function main() {
    // >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<
    const configurator: ExampleTestsuiteConfigurator = new ExampleTestsuiteConfigurator();
    // >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<

    const suiteExecutor: TestSuiteExecutor = new TestSuiteExecutor(configurator);
    let exitCode: number = SUCCESS_EXIT_CODE;
    const suiteExecutorResult: Result<null, Error> = await suiteExecutor.run();
    if (!suiteExecutorResult.isOk()) {
        log.error("An error occurred running the test suite executor:");
        console.log(err);
        exitCode = FAILURE_EXIT_CODE;
    }
    process.exit(exitCode)
}
