export class Config {
    public wsEndpoint: string;
    public httpEndpoint: string;

    constructor(
        public readonly apiHost: string,
        public readonly clientId: string,
        public readonly debug: boolean
    ) {
        let wsProto: string = debug ? 'ws' : 'wss';
        let httpProto: string = debug ? 'http' : 'https';
        this.wsEndpoint = `${wsProto}://${this.apiHost}/ws`;
        this.httpEndpoint = `${httpProto}://${this.apiHost}/api`;
    }
}