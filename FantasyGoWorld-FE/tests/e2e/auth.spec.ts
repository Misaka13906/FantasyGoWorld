import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {

  test('TC-AUTH-P-01: Valid Registration and Redirection to Login', async ({ page }) => {
    const randomSuffix = Math.random().toString(36).substring(7);
    const testUser = `p_${randomSuffix}`;

    await page.goto('/login');
    await page.click('text=还没有账号？去注册');
    
    await page.fill('input[placeholder="请输入用户名"]', testUser);
    await page.fill('input[placeholder="请输入密码"]', '123456');
    await page.fill('input[placeholder="请输入昵称"]', 'Tester');
    await page.selectOption('select', '18K');

    await page.click('button[type="submit"]');

    await expect(page.locator('text=注册成功')).toBeVisible(); 
    await expect(page.locator('button[type="submit"]')).toHaveText('登录');
  });

  test('TC-AUTH-P-02/04: Valid Login, Persistence (Reload), and Logout', async ({ page, context }) => {
    const username = `u_${Math.random().toString(36).substring(7)}`;
    const nickname = 'Persistence King';
    
    // Register
    await page.goto('/login');
    await page.click('text=还没有账号？去注册');
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[placeholder="请输入密码"]', 'password123');
    await page.fill('input[placeholder="请输入昵称"]', nickname);
    await page.click('button[type="submit"]');
    await expect(page.locator('button[type="submit"]')).toHaveText('登录');

    // Login
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[placeholder="请输入密码"]', 'password123');
    await page.click('button[type="submit"]');

    // Assert URL and UI
    await expect(page).toHaveURL(/.*\/lobby/);
    await expect(page.locator('h1')).toContainText('棋幻大厅');
    await expect(page.locator('header')).toContainText('欢迎');
    await expect(page.locator('header')).toContainText(nickname);

    // Verify persistence via reload
    await page.reload();
    await expect(page.locator('header')).toContainText(nickname);
    await expect(page).toHaveURL(/.*\/lobby/);

    // Logout
    await page.click('text=退出登录');
    await expect(page).toHaveURL(/.*\/login/);

    const cookies = await context.cookies();
    const hasAccessToken = cookies.some(c => c.name === 'access_token');
    expect(hasAccessToken).toBe(false);
  });

  test('TC-AUTH-P-03: Route Guard redirects to /login', async ({ page }) => {
    await page.goto('/lobby');
    await expect(page).toHaveURL(/.*\/login/);
  });

  test('TC-AUTH-N-01: Registration with Duplicate Username', async ({ page }) => {
    const username = `dup_${Math.random().toString(36).substring(7)}`;
    
    // First registration
    await page.goto('/login');
    await page.click('text=还没有账号？去注册');
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[placeholder="请输入密码"]', 'password123');
    await page.fill('input[placeholder="请输入昵称"]', 'Player A');
    await page.click('button[type="submit"]');
    await expect(page.locator('button[type="submit"]')).toHaveText('登录');

    // Second registration with same username
    await page.click('text=还没有账号？去注册');
    await page.fill('input[placeholder="请输入用户名"]', username);
    await page.fill('input[placeholder="请输入密码"]', 'password123');
    await page.fill('input[placeholder="请输入昵称"]', 'Player B');
    await page.click('button[type="submit"]');

    // Should show error toast
    await expect(page.locator('text=username already exists')).toBeVisible();
  });

  test('TC-AUTH-N-02/03: Client-side Field Validation (Short Input)', async ({ page }) => {
    await page.goto('/login');
    await page.click('text=还没有账号？去注册');

    // Short username
    await page.fill('input[placeholder="请输入用户名"]', 'abc');
    await page.fill('input[placeholder="请输入密码"]', '123456');
    await page.fill('input[placeholder="请输入昵称"]', 'Shorty');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=用户名至少需要 4 个字符')).toBeVisible();

    // Short password
    await page.fill('input[placeholder="请输入用户名"]', 'validuser');
    await page.fill('input[placeholder="请输入密码"]', '12345');
    await page.click('button[type="submit"]');
    await expect(page.locator('text=密码至少需要 6 个字符')).toBeVisible();
  });

  test('TC-AUTH-N-04: Invalid Login Credentials', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="请输入用户名"]', 'notexistuser');
    await page.fill('input[placeholder="请输入密码"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=invalid credentials')).toBeVisible();
  });

});
