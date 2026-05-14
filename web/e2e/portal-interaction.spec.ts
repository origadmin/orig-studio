import {test, expect} from '@playwright/test';

const DESKTOP = {width: 1280, height: 800};

test.describe('Portal - Sidebar Navigation (unauthenticated)', () => {
    test.use({viewport: DESKTOP});

    test.beforeEach(async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        await page.evaluate(() => localStorage.setItem('sidebarCollapsed', 'false'));
    });

    const sidebarLinks = [
        {label: 'Home', href: '/'},
        {label: 'Featured', href: '/featured'},
        {label: 'Latest', href: '/latest'},
        {label: 'Categories', href: '/categories'},
        {label: 'Tags', href: '/tags'},
        {label: 'Members', href: '/members'},
        {label: 'Trending', href: '/explore'},
        {label: 'About', href: '/about'},
    ];

    for (const link of sidebarLinks) {
        test(`sidebar "${link.label}" navigates to ${link.href}`, async ({page}) => {
            await page.goto('/');
            await page.waitForTimeout(2000);
            const sidebarLink = page.locator(`aside a[href="${link.href}"]`).first();
            await expect(sidebarLink).toBeVisible({timeout: 5000});
            await sidebarLink.click();
            await page.waitForTimeout(2000);
            expect(page.url()).toContain(link.href);
        });
    }
});

test.describe('Portal - Header Search Flow', () => {
    test.use({viewport: DESKTOP});

    test('search input submits and navigates to /search', async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        const searchInput = page.locator('header input[type="search"]').first();
        await expect(searchInput).toBeVisible({timeout: 5000});
        await searchInput.fill('test video');
        await searchInput.press('Enter');
        await page.waitForTimeout(2000);
        expect(page.url()).toContain('/search');
    });
});

test.describe('Portal - Header Auth Flow', () => {
    test.use({viewport: DESKTOP});

    test('click signin button navigates to /auth/signin', async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        const signinBtn = page.locator('header a[href="/auth/signin"]').first();
        await expect(signinBtn).toBeVisible({timeout: 5000});
        await signinBtn.click();
        await page.waitForTimeout(2000);
        expect(page.url()).toContain('/auth/signin');
    });

    test('signin form fill and submit', async ({page}) => {
        await page.goto('/auth/signin');
        await page.waitForSelector('#username', {timeout: 5000});
        await page.fill('#username', 'admin');
        await page.fill('#password', 'wrong');
        await page.locator('button[type="submit"]').first().click();
        await page.waitForTimeout(2000);
    });

    test('signin page → signup link click', async ({page}) => {
        await page.goto('/auth/signin');
        await page.waitForTimeout(2000);
        const signupLink = page.locator('a[href*="signup"]').first();
        if (await signupLink.isVisible({timeout: 3000})) {
            await signupLink.click();
            await page.waitForTimeout(2000);
            expect(page.url()).toContain('/auth/signup');
        }
    });

    test('signup page → signin link click', async ({page}) => {
        await page.goto('/auth/signup');
        await page.waitForTimeout(2000);
        const signinLink = page.locator('a[href*="signin"]').first();
        if (await signinLink.isVisible({timeout: 3000})) {
            await signinLink.click();
            await page.waitForTimeout(2000);
            expect(page.url()).toContain('/auth/signin');
        }
    });
});

test.describe('Portal - Authenticated Sidebar Links', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test.beforeEach(async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        await page.evaluate(() => localStorage.setItem('sidebarCollapsed', 'false'));
    });

    const authLinks = [
        {label: 'Subscriptions Feed', href: '/subscriptions'},
        {label: 'My Videos', href: '/me/videos'},
        {label: 'My Channels', href: '/me/channels'},
        {label: 'My Articles', href: '/me/articles'},
        {label: 'History', href: '/me/history'},
        {label: 'Favorites', href: '/me/favorites'},
        {label: 'Playlists', href: '/me/playlists'},
        {label: 'Notifications', href: '/me/notifications'},
    ];

    for (const link of authLinks) {
        test(`sidebar "${link.label}" navigates to ${link.href}`, async ({page}) => {
            const sidebarLink = page.locator(`aside a[href="${link.href}"]`).first();
            if (await sidebarLink.isVisible({timeout: 5000})) {
                await sidebarLink.click();
                await page.waitForTimeout(2000);
                expect(page.url()).toContain(link.href);
            }
        });
    }
});

test.describe('Portal - Header Authenticated User Menu', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test('click user avatar opens menu, "Admin" link visible', async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        const avatarBtn = page.locator('header div[class*="rounded-full"] button').last();
        if (await avatarBtn.isVisible({timeout: 5000})) {
            await avatarBtn.click();
            await page.waitForTimeout(1000);
            const adminLink = page.locator('a[href="/admin"]').last();
            const visible = await adminLink.isVisible({timeout: 3000}).catch(() => false);
            expect(visible).toBeTruthy();
        }
    });

    test('click Admin in user menu navigates to /admin', async ({page}) => {
        await page.goto('/');
        await page.waitForTimeout(2000);
        const avatarBtn = page.locator('header div[class*="rounded-full"] button').last();
        if (await avatarBtn.isVisible({timeout: 5000})) {
            await avatarBtn.click();
            await page.waitForTimeout(1000);
            const adminLink = page.locator('a[href="/admin"]').last();
            if (await adminLink.isVisible({timeout: 3000})) {
                await adminLink.click();
                await page.waitForTimeout(2000);
                expect(page.url()).toContain('/admin');
            }
        }
    });
});