import {defineConfig, devices} from '@playwright/test';

export default defineConfig({
    testDir: './e2e',
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: 'html',
    timeout: 60000,
    use: {
        baseURL: 'http://localhost:18080',
        trace: 'on-first-retry',
        screenshot: 'on',
    },
    projects: [
        {name: 'setup', testMatch: /auth\.setup\.ts/},
        {
            name: 'chromium-auth',
            use: {
                ...devices['Desktop Chrome'],
                storageState: '.auth/user.json',
            },
            dependencies: ['setup'],
            testMatch: /admin-interaction\.spec\.ts/,
        },
        {
            name: 'chromium',
            use: {...devices['Desktop Chrome']},
            dependencies: ['setup'],
            testMatch: /(portal-interaction|auth)\.spec\.ts/,
        },
    ],
    webServer: {
        command: 'bun run dev',
        url: 'http://localhost:18080',
        reuseExistingServer: !process.env.CI,
        timeout: 120 * 1000,
    },
});