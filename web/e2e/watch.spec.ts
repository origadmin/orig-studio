import {test, expect} from '@playwright/test';

test.describe('视频观看流程', () => {
    test.beforeEach(async ({page}) => {
        // 每个测试前登录
        await page.goto('/login');
        await page.fill('input[name="username"]', 'user1');
        await page.fill('input[name="password"]', 'user123');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/');
    });

    test('观看视频并点赞', async ({page}) => {
        // 访问视频页面
        await page.goto('/watch/1');

        // 等待视频加载
        await expect(page.locator('video, .video-player')).toBeVisible();

        // 验证视频标题
        await expect(page.locator('h1')).toBeVisible();

        // 点击点赞按钮
        const likeButton = page.locator('[data-testid="like-button"], button:has-text("点赞")').first();
        await likeButton.click();

        // 验证点赞状态变化
        await expect(page.locator('[data-testid="liked"], .liked')).toBeVisible();
    });

    test('收藏视频到播放列表', async ({page}) => {
        await page.goto('/watch/1');

        // 点击收藏按钮
        const favButton = page.locator('[data-testid="favorite-button"], button:has-text("收藏")').first();
        await favButton.click();

        // 验证收藏弹窗或提示
        await expect(page.locator('.toast, [role="dialog"], .notification')).toBeVisible();
    });

    test('发表评论', async ({page}) => {
        await page.goto('/watch/1');

        // 填写评论
        const commentInput = page.locator('textarea[placeholder*="评论"], input[placeholder*="评论"]').first();
        await commentInput.fill('这是一个测试评论');

        // 提交评论
        await page.click('button:has-text("发送"), button:has-text("评论")');

        // 验证评论显示
        await expect(page.locator('.comment:has-text("这是一个测试评论")')).toBeVisible();
    });

    test('订阅视频作者', async ({page}) => {
        await page.goto('/watch/1');

        // 找到订阅按钮并点击
        const subscribeBtn = page.locator('[data-testid="subscribe-button"], button:has-text("订阅")').first();
        await subscribeBtn.click();

        // 验证按钮状态变为"已订阅"
        await expect(page.locator('button:has-text("已订阅"), [data-testid="subscribed"]')).toBeVisible();
    });
});

test.describe('视频列表浏览', () => {
    test('首页加载视频列表', async ({page}) => {
        await page.goto('/');

        // 等待视频卡片加载
        await expect(page.locator('.video-card, [data-testid="video-item"]').first()).toBeVisible();

        // 验证有多个视频
        const videos = await page.locator('.video-card, [data-testid="video-item"]').count();
        expect(videos).toBeGreaterThan(0);
    });

    test('搜索视频', async ({page}) => {
        await page.goto('/');

        // 在搜索框输入
        await page.fill('input[type="search"], input[placeholder*="搜索"]', '测试');
        await page.keyboard.press('Enter');

        // 验证搜索结果页面
        await expect(page).toHaveURL(/.*search.*/);

        // 验证有搜索结果或空状态
        await expect(
            page.locator('.video-card, [data-testid="video-item"], .empty-state').first()
        ).toBeVisible();
    });
});
