import type {Config} from 'jest';

const config: Config = {
    testEnvironment: 'jsdom',
    setupFilesAfterEnv: ['<rootDir>/src/test/setup.ts'],
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
        '^@/router$': '<rootDir>/src/router/index.tsx',
        '^@tests/(.*)$': '<rootDir>/tests/$1',
        '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    },
    transform: {
        '^.+\\.tsx?$': ['ts-jest', {
            tsconfig: '<rootDir>/tsconfig.json',
            useESM: false
        }],
    },
    testMatch: [
        '<rootDir>/src/**/*.test.ts',
        '<rootDir>/src/**/*.test.tsx',
        '<rootDir>/tests/**/*.test.ts',
        '<rootDir>/tests/**/*.test.tsx',
    ],
    testTimeout: 10000,
    forceExit: true,
    detectOpenHandles: false,
    clearMocks: true,
    restoreMocks: true,
    transformIgnorePatterns: [
        '/node_modules/(?!(axios|@tanstack)/)',
    ],
};

export default config;
