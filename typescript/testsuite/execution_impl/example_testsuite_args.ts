export class ExampleTestsuiteArgs {
    private readonly apiServiceImage: string;
    private readonly datastoreServiceImage: string;
    
    constructor(apiServiceImage: string, datastoreServiceImage: string) {
        this.apiServiceImage = apiServiceImage;
        this.datastoreServiceImage = datastoreServiceImage;
    }
    
    public getApiServiceImage(): string {
        return this.apiServiceImage;
    }
    
    public getDatastoreServiceImage(): string {
        return this.datastoreServiceImage;
    }
}