import { before, after, test } from 'node:test';
import assert from 'node:assert/strict';
import { readFile, mkdir } from 'node:fs/promises';
import { randomUUID } from 'node:crypto';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { chromium } from 'playwright';

if (!process.env.QMANGA_FIXTURE_FILE) throw Error('QMANGA_FIXTURE_FILE required; no reading browser tests are skipped.');
const fixture = JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE, 'utf8'));
const baseURL = fixture.app_url;
if (!['localhost', '127.0.0.1'].includes(new URL(baseURL).hostname) || !fixture.env_file || !/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name || '')) throw Error('A dedicated qmanga-ci or qmanga-stageN-* fixture is required; never seed the owner environment.');
let browser, owner, ownerId, teamId, otherTeam, libraryTarget, guestRows, coverURL;
const distantChapters = [];
const readingActorID=randomUUID(),readingEmail='reading-'+readingActorID+'@test.local';
const titlePrefix = randomUUID().slice(0, 8);
const titles = Array.from({ length: 110 }, (_, index) => ({ id: titlePrefix + '-0000-4000-8000-' + (110 - index).toString(16).padStart(12, '0'), slug: 'reading-' + randomUUID(), project: randomUUID(), chapter: randomUUID(), release: randomUUID(), revision: randomUUID(), page: randomUUID(), order: index + 1 }));
function sql(text) {
 const args = ['compose', '--env-file', fixture.env_file];
 if (fixture.project_name) args.push('-p', fixture.project_name);
 args.push('exec', '-T', 'postgres', 'psql', '-U', 'qmanga', '-d', 'qmanga', '-v', 'ON_ERROR_STOP=1');
 const result = spawnSync('docker', args, { input: text, encoding: 'utf8', timeout: 30000, windowsHide: true });
 if (result.status !== 0) throw Error('Disposable reading fixture SQL failed: ' + result.stderr);
}
function cleanupSql() {
 const titleIds = titles.map(t => `'${t.id}'`).join(',');
 const distantChapterFilter = distantChapters.length
  ? `ch.id IN (${distantChapters.map(c => `'${c}'`).join(',')})`
  : 'false';
 const distantChapterIdFilter = distantChapters.length
  ? `id IN (${distantChapters.map(c => `'${c}'`).join(',')})`
  : 'false';
 return `BEGIN;
DELETE FROM library_entries WHERE user_id='${ownerId}' AND title_id IN (${titleIds});
DELETE FROM reading_progress WHERE user_id='${ownerId}' AND title_id IN (${titleIds});
DELETE FROM pages WHERE revision_id IN (SELECT rev.id FROM revisions rev JOIN releases rel ON rel.id = rev.release_id JOIN chapters ch ON ch.id = rel.chapter_id WHERE ch.title_id IN (${titleIds}));
DELETE FROM revisions WHERE release_id IN (SELECT rel.id FROM releases rel JOIN chapters ch ON ch.id = rel.chapter_id WHERE ch.title_id IN (${titleIds}));
DELETE FROM releases WHERE chapter_id IN (SELECT id FROM chapters WHERE title_id IN (${titleIds}));
DELETE FROM chapters WHERE title_id IN (${titleIds});
DELETE FROM projects WHERE title_id IN (${titleIds});
DELETE FROM titles WHERE id IN (${titleIds});
DELETE FROM team_memberships WHERE team_id='${otherTeam}';
DELETE FROM teams WHERE id='${otherTeam}';
DELETE FROM revisions WHERE release_id IN (SELECT rel.id FROM releases rel JOIN chapters ch ON ch.id = rel.chapter_id WHERE ${distantChapterFilter});
DELETE FROM releases WHERE chapter_id IN (SELECT id FROM chapters WHERE ${distantChapterIdFilter});
DELETE FROM chapters WHERE ${distantChapterIdFilter};
COMMIT;`;
}
before(async () => {
 browser = await chromium.launch({ headless: true, ...(process.env.QMANGA_BROWSER_EXECUTABLE ? { executablePath: process.env.QMANGA_BROWSER_EXECUTABLE } : {}) });
 owner = await browser.newContext({ baseURL, viewport: { width: 360, height: 800 } });
 await owner.addCookies([{ name: 'qmanga_locale', value: 'ru', url: baseURL }]);
 assert.match(fixture.owner_email,/^[a-zA-Z0-9@._+-]+$/);assert.match(fixture.project_id,/^[a-f0-9-]{36}$/);
 // Pagination totals belong to this test account, independent of bookmarks left by other flows.
 sql(`INSERT INTO users(id,email,password_hash,display_name,is_verified) SELECT '${readingActorID}','${readingEmail}',password_hash,'Reading isolated fixture',true FROM users WHERE email='${fixture.owner_email}';INSERT INTO team_memberships(id,team_id,user_id,role) SELECT '${randomUUID()}',team_id,'${readingActorID}','editor' FROM projects WHERE id='${fixture.project_id}';`);

 const login = await owner.request.post('/api/v1/auth/login', { headers: { Origin: baseURL }, data: { email: readingEmail, password: fixture.owner_password } });
 assert.equal(login.status(), 200);
 ownerId = (await (await owner.request.get('/api/v1/auth/me')).json()).user.id;
 const project = await (await owner.request.get('/api/v1/studio/projects/' + fixture.project_id)).json();
 teamId = project.team_id; otherTeam = randomUUID();
 for (const id of [ownerId, teamId, fixture.title_id, fixture.project_id]) assert.match(id, /^[a-f0-9-]{36}$/);
 const source = await (await owner.request.get('/api/v1/releases/' + fixture.release_id + '/manifest')).json();
 coverURL = source.cover_url;
 libraryTarget = [...titles].sort((left, right) => left.id.localeCompare(right.id))[0];
 let setup = `BEGIN;
INSERT INTO teams(id,slug,name,is_approved)VALUES('${otherTeam}','reading-${otherTeam}','Екінші сынақ командасы',true); INSERT INTO team_memberships(id,team_id,user_id,role)VALUES('${randomUUID()}','${otherTeam}','${ownerId}','owner');`;
 for (const item of titles) {
  const label = 'Ұзақ сынақ ' + String(item.order).padStart(3, '0') + (item === libraryTarget ? ' · Қашық інжу' : '');
  setup += `INSERT INTO titles(id,slug,name_kk,cover_asset_id,is_demo)SELECT '${item.id}','${item.slug}','${label}',cover_asset_id,true FROM titles WHERE id='${fixture.title_id}';
   INSERT INTO projects(id,title_id,team_id)VALUES('${item.project}','${item.id}','${item.order === 110 ? otherTeam : teamId}');
   INSERT INTO chapters(id,title_id,label,order_num)VALUES('${item.chapter}','${item.id}','1',1);
   INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id)VALUES('${item.release}','${item.chapter}','${item.project}','${item.order === 110 ? otherTeam : teamId}','published','${item.revision}');
   INSERT INTO revisions(id,release_id,status)VALUES('${item.revision}','${item.release}','published');
   INSERT INTO pages(id,revision_id,position,width,height,status)VALUES('${item.page}','${item.revision}',1,800,1200,'ready');
   INSERT INTO library_entries(id,user_id,title_id,status,is_favorite)VALUES('${randomUUID()}','${ownerId}','${item.id}','${item === libraryTarget ? 'planned' : 'reading'}',${item === libraryTarget});`;
 }
 for (let number = 1; number <= 305; number++) {
  const chapter = randomUUID(), release = randomUUID(), revision = randomUUID();
  distantChapters.push(chapter);
  setup += `INSERT INTO chapters(id,title_id,label,title_kk,order_num)VALUES('${chapter}','${fixture.title_id}','${number + 1000}','Дала ${String(number).padStart(3, '0')}',${number + 1000});
   INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id)VALUES('${release}','${chapter}','${fixture.project_id}','${teamId}','published','${revision}');
   INSERT INTO revisions(id,release_id,status)VALUES('${revision}','${release}','published');`;
 }
 for (const item of titles.slice(0, 3)) setup += `INSERT INTO reading_progress(id,user_id,title_id,release_id,revision_id,page_id,page_position,fraction,updated_at)VALUES('${randomUUID()}','${ownerId}','${item.id}','${item.release}','${item.revision}','${item.page}',1,0.5,now()+interval '${item.order} minutes');`;
 sql(setup + 'COMMIT;');
 await owner.setExtraHTTPHeaders({ 'X-Qmanga-Account': ownerId });
 guestRows = [...titles].filter(item => !titles.slice(0, 3).includes(item)).sort((left, right) => left.id.localeCompare(right.id)).slice(0, 3);
});
after(async () => {
 try {
  if(ownerId && otherTeam) sql(cleanupSql());
  sql(`DELETE FROM users WHERE id='${readingActorID}';`);
 } finally {
  await owner?.close();
  await browser?.close();
 }
});

test('library and project filters find matches beyond 100 and reset pagination', { timeout: 60000 }, async () => {
 const page = await owner.newPage();
 try {
  await page.goto('/library');
  await page.getByText('Найдено: 110', { exact: true }).waitFor();
  await page.locator('.library-filters').getByRole('button',{name:'В планах',exact:true}).click();
  await page.getByText('Найдено: 1', { exact: true }).waitFor();
  await page.locator(`.card-heading a[href="/titles/${libraryTarget.slug}"]`).waitFor();
  assert.equal(await page.getByRole('button', { name: 'Загрузить ещё', exact: true }).count(), 0);
  await page.locator('.library-card .options').first().click();
  await page.locator('.entry-controls select').first().selectOption('reading');
  await page.getByText('Найдено: 0', { exact: true }).waitFor();
  assert.equal(await page.locator(`a[href="/titles/${libraryTarget.slug}"]`).count(), 0);

  await page.locator('.library-filters').getByRole('button',{name:'Все',exact:true}).click();
  await page.getByText('Найдено: 110', { exact: true }).waitFor();

  assert.equal(await page.locator('.library-search').count(),0);
  await page.getByRole('button', { name: 'Загрузить ещё', exact: true }).click();
  await page.locator(`.card-heading a[href="/titles/${titles[109].slug}"]`).waitFor();
  assert.equal(await page.getByRole('button', { name: 'Загрузить ещё', exact: true }).count(), 0);

  await page.goto('/studio/projects');
  await page.getByLabel('Команда', { exact: true }).selectOption(otherTeam);
  await page.getByText('Найдено: 1', { exact: true }).waitFor();
  await page.locator(`a[href="/studio/projects/${titles[109].project}"]`).first().waitFor();
 } finally { await page.close(); }
});

test('reader searches a distant chapter without loading preceding pages', { timeout: 60000 }, async () => {
 const page = await owner.newPage();
 try {
  await page.goto('/read/' + fixture.release_id);
  await page.locator('.reader-controls button').nth(1).click();
  const dialog = page.getByRole('dialog');
  await dialog.locator('input[type="search"]').fill('Дала 001');
  await dialog.getByRole('button', { name: 'Найти', exact: true }).click();
  await dialog.getByRole('link', { name: /Дала 001/ }).waitFor();
  assert.equal(await dialog.locator('.chapter-list a').count(), 1);
 } finally { await page.close(); }
});

test('account recent history uses activity and no library scan; guest v1 data upgrades and gets its own cover', { timeout: 60000 }, async () => {
 const page = await owner.newPage();
 const libraryRequests = [];
 page.on('request', request => { if (request.url().includes('/api/v1/library?')) libraryRequests.push(request.url()); });
 try {
  await page.goto('/');
  await page.locator('.continuation-grid a').first().waitFor();
  assert.equal(await page.locator('.continuation-grid a').first().getAttribute('href'), '/read/' + titles[2].release);
  assert.equal(libraryRequests.length, 0);
 } finally { await page.close(); }
 const guest = await browser.newContext({ baseURL, viewport: { width: 360, height: 800 }, colorScheme: 'dark' });
 const guestPage = await guest.newPage();
 try {
  await guest.addCookies([{ name: 'qmanga_locale', value: 'ru', url: baseURL }]);
  await guestPage.goto('/catalog');
  await guestPage.evaluate(async ({ rows, cover }) => {
   await new Promise((resolve, reject) => {
    const request = indexedDB.open('qmanga-reading', 1);
    request.onupgradeneeded = () => request.result.createObjectStore('positions', { keyPath: ['scope', 'titleId'] });
    request.onerror = () => reject(request.error);
    request.onsuccess = () => {
     const connection = request.result, work = connection.transaction('positions', 'readwrite');
     for (let index = 0; index < rows.length; index++) {
      const item = rows[index];
      work.objectStore('positions').put({ scope: 'guest', titleId: item.id, slug: item.slug, name: 'Guest ' + item.order, coverURL: cover, updatedAt: 300 - index * 100, pagePosition: 1, progress: { titleId: item.id, releaseId: item.release, revisionId: item.revision, pageId: item.page, fraction: 0.5 }, pending: null, conflict: null });
     }
     work.oncomplete = () => { connection.close(); resolve(); }; work.onerror = () => reject(work.error);
    };
   });
  }, { rows: guestRows, cover: coverURL });
  await guestPage.goto('/');
  const card = guestPage.locator('.continuation-grid a').first();
  await card.waitFor();
  assert.equal(await card.getAttribute('href'), '/read/' + guestRows[0].release);
  assert.equal(await card.locator('img').getAttribute('src'), coverURL);
  if (process.env.QMANGA_SCREENSHOT_DIR) {
   await mkdir(process.env.QMANGA_SCREENSHOT_DIR, { recursive: true });
   await guestPage.screenshot({ path: join(process.env.QMANGA_SCREENSHOT_DIR, 'recent-guest-360.png'), fullPage: true });
  }
 } finally { await guest.close(); }
});

test('library filter error is displayed with retry and does not mask as empty state', { timeout: 60000 }, async () => {
 const page = await owner.newPage();
 try {
  await page.goto('/library');
  await page.getByText('Найдено: 110', { exact: true }).waitFor();

  const failRegex = /\/api\/v1\/library\?.*status=dropped/;
  await page.route(failRegex, route => route.fulfill({
   status: 500,
   contentType: 'application/json',
   body: JSON.stringify({ error: { code: 'server_down', message: 'Internal error' } })
  }));

  await page.locator('.library-filters').getByRole('button',{name:'Брошено',exact:true}).click();

  await page.locator('p[role="alert"]').waitFor();
  assert.equal(await page.locator('.empty-state').count(), 0);

  await page.unroute(failRegex);
  await page.getByRole('button', { name: 'Повторить', exact: true }).click();

  await page.getByText('Найдено: 0', { exact: true }).waitFor();
 } finally {
  await page.close();
 }
});

test('two rapid changes to entry status and favorite do not trigger artificial version conflict', { timeout: 60000 }, async () => {
 const page = await owner.newPage();
 let releaseSave;
 const heldSave = new Promise(resolve => { releaseSave = resolve; });
 let observeSave;
 const saveStarted = new Promise(resolve => { observeSave = resolve; });
 const errors=[];
 page.on('pageerror', error=>errors.push(error.message));
 try {
  await page.goto('/library');
  await page.getByText('Найдено: 110', { exact: true }).waitFor();

  const putPattern = '**/api/v1/library/*';
  await page.route(putPattern, async route => {
   if (route.request().method() === 'PUT') {
    observeSave();
    await heldSave;
   }
   await route.continue();
  });

  await page.locator('.library-card .options').first().click();
  const row = page.locator('.entry-controls').first();
  const select = row.locator('select');
  const checkbox = row.locator('input[type="checkbox"]');

  await select.selectOption('planned');
  await saveStarted;
  assert.equal(await select.isDisabled(), true);
  assert.equal(await checkbox.isDisabled(), true);
  assert.equal(await row.getByRole('button', {name:'Убрать из списка',exact:true}).isDisabled(), true);
  releaseSave();
  await page.getByText('Сохранено', { exact: true }).waitFor();
  await select.selectOption('completed');
  await page.getByText('Сохранено', { exact: true }).waitFor();
  await checkbox.check();

  await page.getByText('Сохранено', { exact: true }).waitFor();

  assert.equal(await page.locator('.conflict').count(), 0);
  assert.equal(await page.getByText('На другом устройстве другой выбор. Какой оставить?', {exact:true}).count(), 0);
  assert.deepEqual(errors,[]);

  await page.unroute(putPattern);
  await page.reload();
  await page.getByText('Найдено: 110', { exact: true }).waitFor();
  await page.locator('.library-card .options').first().click();
  const reloadedRow = page.locator('.entry-controls').first();
  assert.equal(await reloadedRow.locator('select').inputValue(), 'completed');
  assert.equal(await reloadedRow.locator('input[type="checkbox"]').isChecked(), true);
 } finally {
  releaseSave();
  await page.close();
 }
});
