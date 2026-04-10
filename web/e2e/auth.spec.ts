import {test, expect} from '@playwright/test';

test.describe('用户认证流程', () => {
    test('用户登录成功', async ({page}) => {
        // 访问登录页
        await page.goto('/login');

        // 填写登录表单
        await page.fill('input[name="username"]', 'admin');
        await page.fill('input[name="password"]', 'admin123');

        // 点击登录按钮
        await page.click('button[type="submit"]');

        // 等待跳转到首页
        await expect(page).toHaveURL('/');

        // 验证用户菜单显示
        await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();
    });

    test('登录失败显示错误', async ({page}) => {
        await page.goto('/login');

        // 填写错误密码
        await page.fill('input[name="username"]', 'admin');
        await page.fill('input[name="password"]', 'wrongpassword');
        await page.click('button[type="submit"]');

        // 验证错误提示
        await expect(page.locator('.text-red-500, [role="alert"]')).toBeVisible();
    });

    test('未登录用户访问受保护页面被重定向', async ({page}) => {
        // 直接访问需要登录的页面
        await page.goto('/admin');

        // 应该被重定向到登录页
        await expect(page).toHaveURL(/.*login/);
    });
});

test.describe('用户注册流程', () => {
    test('新用户注册成功', async ({page}) => {
        await page.goto('/signup');

        // 填写注册表单
        await page.fill('input[name="username"]', 'newuser' + Date.now());
        await page.fill('input[name="email"]', `new${Date.now()}@example.com`);
        await page.fill('input[name="password"]', 'password123');
        await page.fill('input[name="confirmPassword"]', 'password123');

        await page.click('button[type="submit"]');

        // 注册成功后跳转
        await expect(page).toHaveURL('/');
    });
});
