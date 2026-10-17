import { before, after, test } from 'node:test';
import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { readFile } from 'node:fs/promises';
import { mkdir } from 'node:fs/promises';
import { join } from 'node:path';

if (process.env.QMANGA_FIXTURE_FILE) {
 const fixture = JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE, 'utf8'));
 Object.assign(process.env, { QMANGA_TEST_URL: fixture.app_url, QMANGA_ADMIN_EMAIL: fixture.admin_email, QMANGA_ADMIN_PASSWORD: fixture.admin_password, QMANGA_TITLE_ID: fixture.title_id, QMANGA_PROJECT_ID: fixture.project_id });
}

const baseURL = process.env.QMANGA_TEST_URL || 'http://localhost:8080';
if (!['localhost', '127.0.0.1', '::1'].includes(new URL(baseURL).hostname)) throw new Error('Browser fixtures are local-only.');
for (const name of ['QMANGA_ADMIN_EMAIL', 'QMANGA_ADMIN_PASSWORD', 'QMANGA_TITLE_ID', 'QMANGA_PROJECT_ID']) {
 if (!process.env[name]) throw new Error(`${name} is required; no browser tests are skipped.`);
}
let browser;
before(async () => { browser = await chromium.launch({ headless: true, ...(process.env.QMANGA_BROWSER_EXECUTABLE ? { executablePath: process.env.QMANGA_BROWSER_EXECUTABLE } : {}) }); });
after(async () => { await browser?.close(); });

async function adminPage() {
 const context = await browser.newContext({ baseURL, viewport: { width: 360, height: 800 } });
 await context.addCookies([{ name: 'qmanga_locale', value: 'ru', url: baseURL }]);
 const login = await context.request.post('/api/v1/auth/login', {
  headers: { Origin: baseURL }, data: { email: process.env.QMANGA_ADMIN_EMAIL, password: process.env.QMANGA_ADMIN_PASSWORD }
 });
 assert.equal(login.status(), 200, 'fixture administrator must authenticate');
 return { context, page: await context.newPage() };
}

test('load and selection are not saved; confirmed write is; rejected write retains input', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 try {
  await page.goto('/admin/catalog');
  const picker = page.locator('select').first();
  await picker.locator(`option[value="${process.env.QMANGA_TITLE_ID}"]`).waitFor({ state: 'attached' });
  assert.equal(await page.getByRole('status').count(), 0, 'opening a form is not a write');
  await picker.selectOption(process.env.QMANGA_TITLE_ID);
  const name = page.locator('input[maxlength="200"]').first();
  await page.waitForFunction(() => document.querySelector('input[maxlength="200"]')?.value !== '');
  assert.equal(await page.getByRole('status').count(), 0, 'choosing a title is not a write');
  const original = await name.inputValue();
  const updated = original + ' · UI';
  await name.fill(updated);
  await page.getByRole('button', { name: 'Сохранить тайтл', exact: true }).click();
  await page.getByRole('status').waitFor();
  const saved = await context.request.get(`/api/v1/admin/titles/${process.env.QMANGA_TITLE_ID}`);
  assert.equal((await saved.json()).name, updated, 'success must correspond to persisted data');
  await name.fill('');
  await page.getByRole('button', { name: 'Сохранить тайтл', exact: true }).click();
  assert.equal(await page.getByRole('status').count(), 0, 'native validation must not retain success from the previous write');
  await name.fill(original + ' · unsaved');
  await page.route(`**/api/v1/admin/titles/${process.env.QMANGA_TITLE_ID}`, async route => {
   if (route.request().method() === 'PUT') await route.fulfill({ status: 409, json: { error: { code: 'conflict', message: 'Conflict' } } });
   else await route.continue();
  });
  await page.getByRole('button', { name: 'Сохранить тайтл', exact: true }).click();
  await page.getByRole('alert').waitFor();
  assert.equal(await name.inputValue(), original + ' · unsaved');
  assert.equal(await page.getByRole('status').count(), 0, 'rejected write must clear previous success');
 } finally { await context.close(); }
});

test('admin outside team follows catalogue project link to a read-only card', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 try {
  await page.goto('/admin/catalog');
  const picker = page.locator('select').first();
  await picker.locator(`option[value="${process.env.QMANGA_TITLE_ID}"]`).waitFor({ state: 'attached' });
  await picker.selectOption(process.env.QMANGA_TITLE_ID);
  const link = page.locator(`a[href$="/projects/${process.env.QMANGA_PROJECT_ID}"]`);
  await link.click();
  await page.getByTestId('admin-project').waitFor();
  assert.equal(new URL(page.url()).pathname, `/admin/projects/${process.env.QMANGA_PROJECT_ID}`);
  assert.equal(await page.getByTestId('admin-project').locator('input[type="file"], form').count(), 0, 'administrative read must not become publishing authority');
  assert.equal(await page.getByRole('alert').count(), 0);
  assert.equal(await page.locator('a[href="/moderation"]').count() > 0, true);
  const studio = await context.request.get(`/api/v1/studio/projects/${process.env.QMANGA_PROJECT_ID}`);
  assert.equal(studio.status(), 403, 'staff read must not grant team membership');
  if (process.env.QMANGA_SCREENSHOT_DIR) {
   await mkdir(process.env.QMANGA_SCREENSHOT_DIR, { recursive: true });
   for (const locale of ['kk', 'ru']) {
    await context.addCookies([{ name: 'qmanga_locale', value: locale, url: baseURL }]);
    await page.reload();
    await page.getByTestId('admin-project').waitFor();
    for (const theme of ['light', 'dark']) {
     await page.emulateMedia({ colorScheme: theme, reducedMotion: 'reduce' });
     for (const width of [320, 360, 1280]) {
      await page.setViewportSize({ width, height: 900 });
      const fits = await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth);
      assert.equal(fits, true, `administrative card must fit ${locale}/${theme}/${width}px`);
      const actions = await page.getByTestId('admin-project').locator('a').all();
      for (const action of actions) assert.equal((await action.boundingBox()).height >= 44, true, 'card actions must keep 44px targets');
      await page.screenshot({ path: join(process.env.QMANGA_SCREENSHOT_DIR, `admin-project-${locale}-${theme}-${width}.png`), fullPage: true });
     }
    }
   }
  }
 } finally { await context.close(); }
});
