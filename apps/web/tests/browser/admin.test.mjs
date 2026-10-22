import { before, after, test } from 'node:test';
import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { readFile } from 'node:fs/promises';
import { randomUUID } from 'node:crypto';
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
  await page.goto('/admin/catalog');await page.getByRole('button',{name:'Создать тайтл',exact:true}).click();
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

test('admin tag picker stays closed after a mouse pick and chip removal', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 const suffix = randomUUID().slice(0, 8), name = 'Тег ' + suffix, slug = 'picker-' + suffix;
 let headers, genre;
 try {
  headers = { Origin: baseURL, 'X-CSRF-Token': (await (await context.request.get('/api/v1/auth/csrf')).json()).csrf_token };
  const created = await context.request.post('/api/v1/admin/genres', { headers, data: { name, slug } });
  assert.equal(created.status(), 201);
  genre = await created.json();
  await page.goto('/admin/catalog?title=' + process.env.QMANGA_TITLE_ID);
  const input = page.getByRole('combobox', { name: 'Теги', exact: true });
  await input.waitFor();
  await input.fill(name);
  const option = page.getByRole('option', { name, exact: true });
  await option.waitFor();
  await option.click();
  assert.equal(await page.getByRole('listbox').count(), 0, 'a mouse pick must not reopen the suggestions');
  assert.equal(await input.getAttribute('aria-expanded'), 'false');
  assert.equal(await input.evaluate(element => element === document.activeElement), true, 'focus returns to the search input');
  const chip = page.getByRole('button', { name, exact: true });
  await chip.waitFor();
  await chip.click();
  assert.equal(await page.getByRole('listbox').count(), 0, 'chip removal must not reopen the suggestions');
  assert.equal(await input.getAttribute('aria-expanded'), 'false');
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true, 'the admin picker fits 360px');
 } finally {
  if (headers && genre) await context.request.delete('/api/v1/admin/genres/' + genre.id, { headers });
  await context.close();
 }
});

test('admin outside team follows catalogue project link to a read-only card', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 try {
  await page.goto('/admin/catalog');await page.getByRole('button',{name:'Создать тайтл',exact:true}).click();
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

test('creating a project preserves unsaved title metadata and category choices', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 try {
  await page.goto('/admin/catalog');await page.getByRole('button',{name:'Создать тайтл',exact:true}).click();
  const picker=page.locator('select').first();
  await picker.locator(`option[value="${process.env.QMANGA_TITLE_ID}"]`).waitFor({state:'attached'});
  await picker.selectOption(process.env.QMANGA_TITLE_ID);
  const name=page.locator('input[maxlength="200"]').first();
  await page.waitForFunction(()=>document.querySelector('input[maxlength="200"]')?.value!=='');
  const draft=(await name.inputValue())+' · draft before project';
  await name.fill(draft);
  const aliases=page.locator('textarea[rows="3"]').first();
  await aliases.fill('Unsaved alias');
  const authors=page.locator('textarea[maxlength="2000"]').first();
  await authors.fill('Unsaved author');
  const category=page.locator('fieldset').filter({has:page.locator('legend', {hasText:'Категории содержания'})}).locator('input[type="checkbox"]').first();
  let checked;
  if(await category.count()){checked=!(await category.isChecked());await category.setChecked(checked);}
  await page.route(`**/api/v1/admin/titles/${process.env.QMANGA_TITLE_ID}/projects`,route=>route.fulfill({status:201,json:{id:'ui-only-project'}}));
  const projectForm=page.locator('form').filter({has:page.getByRole('button',{name:'Создать проект',exact:true})});
  const team=projectForm.locator('select');
  const value=await team.locator('option').evaluateAll(options=>options.find(option=>option.value)?.value);
  assert.ok(value,'fixture must provide an approved team');
  await team.selectOption(value);
  await projectForm.getByRole('button',{name:'Создать проект',exact:true}).click();
  await page.getByRole('status').waitFor();
  assert.equal(await name.inputValue(),draft,'creating a project must not replace unsaved metadata');
  assert.equal(await aliases.inputValue(),'Unsaved alias');
  assert.equal(await authors.inputValue(),'Unsaved author');
  if(await category.count())assert.equal(await category.isChecked(),checked);
 } finally {await context.close();}
});

test('focused reply beyond the first batch does not skip intermediate replies on load more', { timeout: 60000 }, async () => {
 const { context, page } = await adminPage();
 try {
  const response=await context.request.get(`/api/v1/admin/titles/${process.env.QMANGA_TITLE_ID}`);
  const title=await response.json();
  const make=(number)=>({id:`paging-${number}`,user_id:'fixture-user',display_name:'Reader',parent_id:'paging-root',root_id:'paging-root',body:`Paging reply ${number}`,spoiler:false,hidden:false,created_at:new Date(Date.UTC(2026,0,1,0,0,number)).toISOString(),replies:0});
  const root={...make(0),id:'paging-root',parent_id:'',root_id:'',body:'Paging root',replies:30,reply_cursor:'paging-10'};
  await page.route('**/api/v1/community/titles/*',route=>route.fulfill({json:{comments:[root,...Array.from({length:10},(_,index)=>make(index+1)),make(25)],next:'',score:0,ratings:0}}));
  let requestedCursor;
  await page.route('**/api/v1/community/comments/paging-root/replies?*',route=>{requestedCursor=new URL(route.request().url()).searchParams.get('after');return route.fulfill({json:{comments:Array.from({length:20},(_,index)=>make(index+11)),next:'paging-30'}});});
  await page.goto(`/titles/${encodeURIComponent(title.slug)}?tab=comments&comment=paging-25`);
  await page.locator('#comment-paging-25').waitFor();
  await page.getByRole('button',{name:'Показать остальные ответы'}).click();
  await page.locator('#comment-paging-30').waitFor();
  assert.equal(requestedCursor,'paging-10','focused reply25 must not replace contiguous cursor10');
  for(let number=1;number<=30;number++)assert.equal(await page.locator(`#comment-paging-${number}`).count(),1,`reply${number} must appear exactly once`);
 } finally {await context.close();}
});
