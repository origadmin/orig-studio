import type {Config} from './types';
import {API_BASE_URL, API_PREFIX, REQUEST_TIMEOUT} from '@/lib/request';

const config: Config = {
    api: {
        baseUrl: API_BASE_URL,
        prefix: API_PREFIX,
        timeout: REQUEST_TIMEOUT,
    },
    app: {
        name: 'OrigCMS',
        version: '1.0.0',
    },
};

export default config;
