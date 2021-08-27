export interface ExampleTestsuiteArgs {
    readonly apiServiceImage: string;
    readonly datastoreServiceImage: string;
    
    //TODO - Remove if everything works
    // constructor(apiServiceImage: string, datastoreServiceImage: string) {
    //     this.apiServiceImage = apiServiceImage;
    //     this.datastoreServiceImage = datastoreServiceImage;
    // }
    
    // public getApiServiceImage(): string {
    //     return this.apiServiceImage;
    // }
    
    // public getDatastoreServiceImage(): string {
    //     return this.datastoreServiceImage;
    // }
}