import {test as setup, expect} from '@playwright/test';
import path from 'path';
import {fileURLToPath} from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const authFile = path.join(__dirname, '..', '.auth', 'user.json');

setup('authenticate as admin', async ({page}) => {
    await page.goto('/auth/signin');
    await page.waitForSelector('#username', {timeout: 10000});
    await page.fill('#username', 'admin');
    await page.fill('#password', 'admin123');
    await page.locator('button[type="submit"]').first().click();
    await page.waitForTimeout(5000);

    let token = await page.evaluate(() => localStorage.getItem('origstudio_token'));

    if (!token) {
        console.log('Admin user not found, registering new admin via signup...');

        await page.goto('/auth/signup');
        await page.waitForSelector('#username', {timeout: 10000});
        await page.fill('#username', 'admin');
        await page.fill('#email', 'admin@origadmin.local');
        await page.fill('#password', 'admin123');
        await page.fill('#confirmPassword', 'admin123');
        await page.locator('button[type="submit"]').click();
        await page.waitForTimeout(5000);

        token = await page.evaluate(() => localStorage.getItem('origstudio_token'));

        if (!token) {
            console.log('Signup done but no auto-login token, trying signin...');
            await page.goto('/auth/signin');
            await page.waitForSelector('#username', {timeout: 10000});
            await page.fill('#username', 'admin');
            await page.fill('#password', 'admin123');
            await page.locator('button[type="submit"]').first().click();
            await page.waitForTimeout(5000);
            token = await page.evaluate(() => localStorage.getItem('origstudio_token'));
        }
    }

    console.log(`Token after auth: ${token ? 'PRESENT' : 'MISSING'}`);
    expect(token).toBeTruthy();

    await page.context().storageState({path: authFile});
});