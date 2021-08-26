import { AdvancedNetworkTest } from "./advanced_network_test/advanced_network_test";
import { BasicDatastoreAndApiTest } from "./basic_datastore_and_api_test/basic_datastore_and_api_test";
import { BasicDatastoreTest } from "./basic_datastore_test/basic_datastore_test";
import { Test } from "kurtosis-testsuite-api-lib";


export class ExampleTestsuite {
    private readonly apiServiceImage: string;
    private readonly datastoreServiceImage: string;

    constructor(apiServiceImage: string, datastoreServiceImage: string) {
        this.apiServiceImage = apiServiceImage;
        this.datastoreServiceImage = datastoreServiceImage;
    }

    public getTests(): Map<string, Test> {
        const tests: Map<string, Test> = new Map();
        tests.set("basicDatastoreTest", new BasicDatastoreTest(this.datastoreServiceImage));
        tests.set("basicDatastoreAndApiTest", new BasicDatastoreAndApiTest(this.datastoreServiceImage, this.apiServiceImage));
        tests.set("advancedNetworkTest", new AdvancedNetworkTest(this.datastoreServiceImage, this.apiServiceImage));
        return tests;
    }
}
