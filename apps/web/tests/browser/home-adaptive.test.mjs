import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
// Guest reading writes only this disposable browser context's local progress.
test('home prioritizes real local continuation without remounting its loader',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:844}});
 try{const page=await context.newPage();await page.goto('/');await page.waitForLoadState('networkidle');
  const headings=()=>page.locator('.discovery-home>section').evaluateAll(es=>es.map(e=>e.getAttribute('aria-labelledby')));
  assert.equal((await headings())[0],'browse-heading','fresh guest starts with discovery');
  await page.goto('/titles/'+fixture.title_slug);await page.waitForLoadState('networkidle');const releaseHref=await page.locator('.chapter-list a[href^="/read/"]').first().getAttribute('href');assert.ok(releaseHref);await page.goto(releaseHref);await page.waitForLoadState('networkidle');
  await page.locator('.reader-controls button').last().click();
  await page.waitForFunction(()=>new Promise(resolve=>{const request=indexedDB.open('qmanga-reading',2);request.onerror=()=>resolve(false);request.onsuccess=()=>{const db=request.result;if(!db.objectStoreNames.contains('positions')){db.close();resolve(false);return;}const work=db.transaction('positions','readonly'),read=work.objectStore('positions').getAll();read.onsuccess=()=>resolve(read.result.some(row=>row.scope==='guest'&&row.pagePosition>=2));work.oncomplete=()=>db.close();};}));
  let references=0;page.on('request',r=>{if(r.url().includes('/api/v1/reading/references?'))references++;});
  await page.goto('/');await page.locator('.continue-card').waitFor();await page.waitForLoadState('networkidle');
  assert.equal((await headings())[0],'continue-heading');assert.equal(references,1,'reordering must not restart the loader');
  for(const width of [320,360,768,1280]){await page.setViewportSize({width,height:900});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false);}
  await page.goto('/?top=manhua#tops');await page.locator('#tops').waitFor();assert.equal(await page.locator('.shelf-tabs a[aria-current]').getAttribute('href'),'/?top=manhua#tops');
 }finally{await context.close();await browser.close();}
});
