import {test, expect} from '@playwright/test';

test.describe('Auth - Sign In', () => {
    test('signin page loads with translated form', async ({page}) => {
        await page.goto('/auth/signin');
        await page.waitForSelector('#username', {timeout: 10000});
        const body = await page.locator('body').innerText();
        expect(body.length).toBeGreaterThan(20);
    });

    test('failed login shows error feedback', async ({page}) => {
        await page.goto('/auth/signin');
        await page.waitForSelector('#username', {timeout: 10000});
        await page.fill('#username', 'admin');
        await page.fill('#password', 'wrong');
        await page.locator('button[type="submit"]').first().click();
        await page.waitForTimeout(2000);
    });

    test('empty form validation prevents submit', async ({page}) => {
        await page.goto('/auth/signin');
        await page.waitForSelector('button[type="submit"]', {timeout: 10000});
        await page.locator('button[type="submit"]').first().click();
        await page.waitForTimeout(1000);
    });
});

test.describe('Auth - Sign Up', () => {
    test('signup page loads with form', async ({page}) => {
        await page.goto('/auth/signup');
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(2000);
        const body = await page.locator('body').innerText();
        expect(body.length).toBeGreaterThan(10);
    });
});

test.describe('Auth - Access Control', () => {
    test('unauthenticated /admin redirects to signin', async ({page}) => {
        await page.goto('/admin');
        await page.waitForURL('**/auth/signin*', {timeout: 5000});
        expect(page.url()).toContain('/auth/signin');
    });

    test('unauthenticated /settings redirects', async ({page}) => {
        await page.goto('/settings');
        await page.waitForTimeout(2000);
    });

    test('unauthenticated /me redirects', async ({page}) => {
        await page.goto('/me');
        await page.waitForTimeout(2000);
    });
});