import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {

  test('Valid Registration and Redirection to Login', async ({ page }) => {
    // Generate random username to avoid backend duplicate error.
    const randomSuffix = Math.random().toString(36).substring(7);
    const testUser = `player_${randomSuffix}`;

    await page.goto('/login');

    // Attempt to register
    await page.click('text=还没有账号？去注册');
    await page.waitForSelector('text=昵称');

    await page.fill('input[type="text"]:first-of-type', testUser);
    await page.fill('input[type="password"]', '123456');
    await page.fill('input[placeholder="请输入昵称"]', 'TestAgent');
    await page.selectOption('select', '18K');

    await page.click('button[type="submit"]', { force: true });

    // Expecting to see success toast, and be redirected back to the login form
    await expect(page.locator('.go3958317564')).toContainText('注册成功'); // hot-toast class, alternative is text content
    await expect(page.locator('button[type="submit"]')).toHaveText('登录');
  });

  test('Valid Login, Store Cookie, and Route to Lobby', async ({ page, context }) => {
    // We already have go:go in DB or from our test cases, but to be sure we create one or log in with it.
    // The previous test may fail occasionally if random name has conflict, so we log in with an existing user.
    // Assuming backend auto setups test user, but let's just create one here.
    
    // First, register a specific user for this session
    const username = `u_${Math.random().toString(36).substring(7)}`;
    await page.goto('/login');
    await page.click('text=还没有账号？去注册');
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[type="password"]', 'password123');
    await page.fill('input[placeholder="请输入昵称"]', 'Player 1');
    await page.click('button[type="submit"]', { force: true });
    await expect(page.locator('button[type="submit"]')).toHaveText('登录');

    // Then log in with that user
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]', { force: true });

    // Assert URL is /lobby
    await expect(page).toHaveURL(/.*\/lobby/);
    await expect(page.locator('h1')).toContainText('大厅 (施工中...)');
    await expect(page.locator('body')).toContainText('Player 1');

    // Verify auth cookies are set
    const cookies = await context.cookies();
    const hasAccessToken = cookies.some(c => c.name === 'access_token');
    expect(hasAccessToken).toBe(true);

    // Click logout
    await page.click('text=退出登录');
    await expect(page).toHaveURL(/.*\/login/);

    const updatedCookies = await context.cookies();
    const hasAccessTokensAfterLogout = updatedCookies.some(c => c.name === 'access_token');
    expect(hasAccessTokensAfterLogout).toBe(false);
  });

  test('Route Guard redirects to /login', async ({ page }) => {
    await page.goto('/lobby');
    await expect(page).toHaveURL(/.*\/login/);
  });
});
