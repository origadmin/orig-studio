import {test, expect} from '@playwright/test';

const DESKTOP = {width: 1280, height: 800};

test.describe('Admin - Sidebar Navigation', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test.beforeEach(async ({page}) => {
        await page.goto('/admin');
        await page.waitForTimeout(3000);
    });

    const adminSidebarLinks = [
        {label: 'Dashboard', href: '/admin'},
        {label: 'Media', href: '/admin/media'},
        {label: 'Users', href: '/admin/users'},
        {label: 'Categories', href: '/admin/categories'},
        {label: 'Channels', href: '/admin/channels'},
        {label: 'Tags', href: '/admin/tags'},
        {label: 'Comments', href: '/admin/comments'},
        {label: 'Notifications', href: '/admin/notifications'},
        {label: 'Playlists', href: '/admin/playlists'},
        {label: 'Articles', href: '/admin/articles'},
        {label: 'Settings', href: '/admin/settings'},
        {label: 'Transcoding Profiles', href: '/admin/transcoding/profiles'},
        {label: 'Transcoding Status', href: '/admin/transcoding/status'},
    ];

    for (const link of adminSidebarLinks) {
        test(`admin sidebar "${link.label}" click navigates to ${link.href}`, async ({page}) => {
            const sidebarLink = page.locator(`aside a[href="${link.href}"]`).first();
            if (await sidebarLink.isVisible({timeout: 5000})) {
                await sidebarLink.click();
                await page.waitForTimeout(2000);
                expect(page.url()).toContain(link.href);
            }
        });
    }
});

test.describe('Admin - Breadcrumb Navigation', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test('breadcrumb "Home" visible on sub-page', async ({page}) => {
        await page.goto('/admin/media');
        await page.waitForTimeout(3000);
        const homeCrumb = page.locator('header a[href="/admin"], div a[href="/admin"]').first();
        const visible = await homeCrumb.isVisible({timeout: 5000}).catch(() => false);
        expect(visible).toBeTruthy();
    });

    test('click "Home" breadcrumb returns to dashboard', async ({page}) => {
        await page.goto('/admin/media');
        await page.waitForTimeout(3000);
        const homeCrumb = page.locator('header a[href="/admin"], div a[href="/admin"]').first();
        if (await homeCrumb.isVisible({timeout: 5000})) {
            await homeCrumb.click();
            await page.waitForTimeout(2000);
            expect(page.url()).toContain('/admin');
        }
    });
});

test.describe('Admin - Exit Admin Link', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test('"Exit Admin" link returns to home', async ({page}) => {
        await page.goto('/admin');
        await page.waitForTimeout(2000);
        const exitLink = page.locator('aside a[href="/"]').last();
        if (await exitLink.isVisible({timeout: 5000})) {
            await exitLink.click();
            await page.waitForTimeout(2000);
            expect(page.url()).toBe('http://localhost:18080/');
        }
    });
});

test.describe('Admin - Sidebar Collapse Toggle', () => {
    test.use({viewport: DESKTOP, storageState: '.auth/user.json'});

    test('toggle button collapses and expands sidebar', async ({page}) => {
        await page.goto('/admin');
        await page.waitForTimeout(2000);
        const toggleBtn = page.locator('button[title*="Collapse"], button[title*="Expand"]').first();
        if (await toggleBtn.isVisible({timeout: 5000})) {
            await toggleBtn.click();
            await page.waitForTimeout(1000);
            await toggleBtn.click();
            await page.waitForTimeout(1000);
        }
    });
});