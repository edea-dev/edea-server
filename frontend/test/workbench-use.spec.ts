import { test, expect, Page, PlaywrightTestConfig } from '@playwright/test';

const config: PlaywrightTestConfig = {
  // Concise 'dot' for CI, default 'list' when running locally
  reporter: process.env.CI ? 'dot' : 'list',
};

export default config;

// take host to test against from env
let edea_url = process.env.TEST_HOST
if (!edea_url) {
    edea_url = "http://localhost:3000"
}

console.log(`testing: ${edea_url}`);

test.describe.serial('user workflow - alice', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto(edea_url);
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")

        await page.click('text=Login');

        await page.fill('#user', 'alice');
        await page.fill('#password', 'alicealice');
        await page.click('text=Submit');

        const buffer = await page.screenshot();
        console.log(buffer.toString('base64'));

        const logout = page.locator('a[href="/logout"]');
        await expect(logout).toHaveText("Logout");
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('new module (LDO)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'NCP1117 3V3 LDO');
        await page.fill('#sub', '3v3ldo');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules');
        await page.fill('#description', 'A simple 3.3-V LDO');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("NCP1117");
    });

    test('new module (5vpol)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'TPS62135 5V PoL');
        await page.fill('#sub', '5vpol');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules');
        await page.fill('#description', '16-V Input 3-A Output Point-of-Load with TPS62135');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("TPS62135");
    });

    test('new module (GD32)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'GD32E103CBT6');
        await page.fill('#sub', 'GD32E103CBT6');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules');
        await page.fill('#description', 'GD32E103CBT6 based module with required passives');
        await page.selectOption('#category', { label: 'MCU' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("GD32E103CBT6");
    });

    test('new module (USB-C)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'USB-C');
        await page.fill('#sub', 'USB-C');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules');
        await page.fill('#description', 'A simple USB-C module');
        await page.selectOption('#category', { label: 'Connector' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("USB-C");
    });
});

test.describe.serial('user workflow - bob', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto(edea_url);
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")

        await page.click('text=Login');

        await page.fill('#user', 'bob');
        await page.fill('#password', 'bob');
        await page.click('text=Submit');

        const logout = page.locator('a[href="/logout"]');
        await expect(logout).toHaveText("Logout");
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('new module (HT7533)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'HT7533 3V3 LDO');
        await page.fill('#sub', 'HT7533-a');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules');
        await page.fill('#description', 'Holtek HT7533, a cheap LDO for 3V3');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("HT7533");
    });
});
