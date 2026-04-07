export interface Config {
    api: {
        baseUrl: string;
        prefix: string;
        timeout: number;
    };
    app: {
        name: string;
        version: string;
    };
}
