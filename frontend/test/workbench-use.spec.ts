import { test, expect, Page, PlaywrightTestConfig } from '@playwright/test';

const config: PlaywrightTestConfig = {
    // Concise 'dot' for CI, default 'list' when running locally
    reporter: process.env.CI ? 'dot' : 'list',
    timeout: 10000,
};

export default config;

// take host to test against from env
let edea_url = process.env.TEST_HOST
if (!edea_url) {
    edea_url = "https://test.edea.dev"
}

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
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', 'A simple 3.3-V LDO');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("NCP1117");
    });

    test('new module (5vpol)', async () => {
        await page.goto(edea_url);
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'TPS62135 5V PoL');
        await page.fill('#sub', '5vpol');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', '16-V Input 3-A Output Point-of-Load with TPS62135');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("TPS62135");

        await page.click('text=Actions');
        await page.click('text="Build Book"');
        const book = page.locator('id=content').first();
        await expect(book).toContainText("Lorem ipsum");
    });

    test('new module (GD32)', async () => {
        await page.goto(edea_url);
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'GD32E103CBT6');
        await page.fill('#sub', 'GD32E103CBT6');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', 'GD32E103CBT6 based module with required passives');
        await page.selectOption('#category', { label: 'MCU' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("GD32E103CBT6");

        // test viewing a specific revision of it
        await page.click('text=History');
        await page.locator('a[href*="7ada1b90dfda4a00a13df760bff05dc5bb22d95f"] >> text=View Module').click();

        const historic_module_page = page.locator('p:has-text("testing the revision picker")');
        await expect(historic_module_page).toContainText("testing the revision picker");
    });

    test('new module (USB-C)', async () => {
        await page.goto(edea_url);
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'USB-C');
        await page.fill('#sub', 'USB-C');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', 'A simple USB-C module');
        await page.selectOption('#category', { label: 'Connector' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("USB-C");
    });

    test('search for modules', async () => {
        await page.fill('#search', 'gd32');
        await page.click('button[type=submit]');

        const module_page = page.locator('#hits-table').first();
        await expect(module_page).toContainText("alice");
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
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', 'Holtek HT7533, a cheap LDO for 3V3');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("HT7533");
    });

    test('new module (NCV68261)', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/new"]');

        await page.fill('#name', 'NCV68261 RP-Protector');
        await page.fill('#sub', 'NCV68261');
        await page.fill('#repourl', 'https://gitlab.com/edea-dev/test-modules.git');
        await page.fill('#description', 'OnSemi NCV68261 Ideal Diode and High Side Switch NMOS Controller');
        await page.selectOption('#category', { label: 'Power' });

        await page.click('text=Submit');

        const module_page = page.locator('h1').first();
        await expect(module_page).toContainText("NCV68261");
    });
});

test.describe.serial('visitor workflow - docs', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto(edea_url);
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('view 5V PoL module docs', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/explore"]');

        const card = page.locator('a:below(:has-text("TPS62135"))').locator('text=View').first();
        await card.click();

        await page.click('text=Docs');

        const book = page.locator('id=content').first();
        await expect(book).toContainText("Lorem ipsum");
    });
});

test.describe.serial('visitor workflow - parametric search', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto(edea_url);
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('use parametric search', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/search"]');

        await page.locator(`select[name='filterf_u_arch']`).selectOption('ARM Cortex-M3');

        const apply_btn = page.locator('#filter_apply_btn');
        await apply_btn.click();

        // button should be disabled after pressing apply
        await expect(apply_btn).toBeDisabled();

        // check if we have at least one result card
        const results = page.locator('div.search-result > .card');
        await expect(results).not.toBeEmpty();
    });
});

test.describe.serial('visitor workflow - view plot diff', () => {
    let page: Page;

    test.beforeAll(async ({ browser }) => {
        page = await browser.newPage();

        await page.goto(edea_url);
        const logo = page.locator('.navbar-brand');
        await expect(logo).toHaveAttribute("aria-label", "EDeA")
    });

    test.afterAll(async () => {
        await page.close();
    });

    test('view NCV68261 git history', async () => {
        await page.click('#navbarModulesDD');
        await page.click('a[href="/module/explore"]');

        const card = page.locator('a:below(:has-text("NCV68261"))').locator('text=View').first();
        await card.click();

        await page.click('text=History');

        const diffLink = page.locator('text=Diff with HEAD').last();
        await diffLink.click();

        const images = page.locator('img');

        // expect a lot of images to be there, TODO: think of something better
        expect(await images.evaluateAll((imgs, min) => imgs.length >= min, 10));
    });
});
