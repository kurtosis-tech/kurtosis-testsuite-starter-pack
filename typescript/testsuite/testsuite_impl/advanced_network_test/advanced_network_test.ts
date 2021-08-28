import { NetworkContext, Network } from "kurtosis-core-api-lib";
import { TestNetwork } from "../../network_impl/test_network"
import { TestConfigurationBuilder } from "kurtosis-testsuite-api-lib";
import * as log from "loglevel";
import { ok, err, Result } from "neverthrow";
import { APIClient, Person } from "../../api/api_service_client/api_client";

const TEST_PERSON_ID: number = 46;

export class AdvancedNetworkTest {
    private readonly datastoreServiceImage: string;
    private readonly apiServiceImage: string;
    
    constructor (datastoreServiceImage: string, apiServiceImage: string) {
        this.datastoreServiceImage = datastoreServiceImage; 
        this.apiServiceImage = apiServiceImage;
    }

    public configure(builder: TestConfigurationBuilder): void {
        builder.withSetupTimeoutSeconds(60).withRunTimeoutSeconds(60);
    }

    public async setup(networkCtx: NetworkContext): Promise<Result<Network, Error>> {
        const network: TestNetwork = new TestNetwork(networkCtx, this.datastoreServiceImage, this.apiServiceImage);
        // Note how setup logic has been pushed into a custom Network implementation, to make test-writing easy
        
        const setupDatastoreAndTwoApisResult: Result<null, Error> = await network.setupDatastoreAndTwoApis();
        if (!setupDatastoreAndTwoApisResult.isOk()) {
            return err(setupDatastoreAndTwoApisResult.error);
        }
        return ok(network);
    }

    public async run(network: Network): Promise<Result<null, Error>> {
        // TODO remove this when test is generic - right now we have to do this cast first
        const castedNetwork: TestNetwork = <TestNetwork>network;
        const personModifierClientResult: Result<APIClient, Error> = castedNetwork.getPersonModifyingApiClient()
        if (!personModifierClientResult.isOk()) {
            return err(personModifierClientResult.error);
        }
        const personModifierClient: APIClient = personModifierClientResult.value;

        const personRetrieverClientResult: Result<APIClient, Error> = castedNetwork.getPersonRetrievingApiClient()
        if (!personRetrieverClientResult.isOk()) {
            return err(personRetrieverClientResult.error);
        }
        const personRetrieverClient: APIClient = personRetrieverClientResult.value;

        log.info("Adding test person via person-modifying API client...");
        const addPersonResult: Result<null, Error> = await personModifierClient.addPerson(TEST_PERSON_ID);;
        if (!addPersonResult.isOk()) {
            return err(addPersonResult.error);
        }
        log.info("Test person added");

        log.info("Incrementing test person's number of books read through person-modifying API client...");
        const incrementBooksReadResult: Result<null, Error> = await personModifierClient.incrementBooksRead(TEST_PERSON_ID);
        if (!incrementBooksReadResult.isOk()) {
            return err(incrementBooksReadResult.error);
        }
        log.info("Incremented number of books read");

        log.info("Retrieving test person to verify number of books read by the person-retrieving API client...");
        const getPersonResult: Result<Person, Error> = await personRetrieverClient.getPerson(TEST_PERSON_ID);
        if (!getPersonResult.isOk()) {
            return err(getPersonResult.error);
        }
        const person: Person = getPersonResult.value;
        log.info("Retrieved test person");

        if (person.booksRead !== 1) {
            return err(new Error("Expected number of books read to be incremented, but was '"+ person.booksRead +"'"));
        }
        return ok(null);
    }
}