import { test, expect, Page } from '@playwright/test';

test.describe.serial('login to merged project download', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto('http://localhost:3000/');
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")

        await page.click('text=Login');

        await page.fill('#user', 'alice');
        await page.fill('#password', 'alicealice');
        await page.click('text=Submit');

        const logout = page.locator('a[href="/logout"]');
        await expect(logout).toHaveText("Logout");
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('browse modules', async () => {
        // page is signed in.
    });

    test('add modules to bench', async () => {
        // page is signed in.  
    });
});
