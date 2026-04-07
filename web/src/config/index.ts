import type {Config} from './types';

const config: Config = {
    api: {
        baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090',
        prefix: '/api/v1',
        timeout: 30000,
    },
    app: {
        name: 'OrigCMS',
        version: '1.0.0',
    },
};

export default config;
