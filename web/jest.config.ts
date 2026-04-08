import type {Config} from 'jest';

const config: Config = {
    testEnvironment: 'jsdom',
    setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
        '^@/router$': '<rootDir>/src/router/index.tsx',
        '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    },
    transform: {
        '^.+\\.tsx?$': ['ts-jest', {
            tsconfig: '<rootDir>/tsconfig.json',
            useESM: false
        }],
    },
    testMatch: ['**/*.test.ts', '**/*.test.tsx'],
    testTimeout: 10000,
    forceExit: true,
    detectOpenHandles: false,
    clearMocks: true,
    restoreMocks: true,
    // 转换包含 import.meta 的模块
    transformIgnorePatterns: [
        '/node_modules/(?!(axios|@tanstack)/)',
    ],
};

export default config;