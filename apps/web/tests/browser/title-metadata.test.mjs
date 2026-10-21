import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('title counts match the API, combined information reflows and icon sorting preserves chapter queries',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
   const context=await browser.newContext({baseURL:fixture.app_url});
   try{
    await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
    const data=await (await context.request.get('/api/v1/titles/'+(fixture.title_slug||fixture.title_id))).json();assert.equal(typeof data.comment_count,'number');
    const page=await context.newPage();page.setDefaultTimeout(7000);await page.goto('/titles/'+data.slug+'?tab=about');await page.waitForLoadState('networkidle');
    assert.equal(await page.locator('.title-tabs a[href*="tab=chapters"] .tab-count').textContent(),String(data.chapter_count));
    assert.equal(await page.locator('.title-tabs a[href*="tab=comments"] .tab-count').textContent(),String(data.comment_count));
    for(const width of [320,360,1280]){
     await page.setViewportSize({width,height:850});
     assert.equal(await page.locator('.metadata').count(),1);
     assert.ok(await page.locator('.title-information').evaluate(e=>e.scrollWidth<=e.clientWidth));
     assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
     if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/title-${locale}-${theme}-${width}.png`});}
    }
    await page.locator('.title-tabs a[href*="tab=chapters"]').click();await page.locator('.chapter-search').waitFor();
    assert.equal(await page.locator('.title-information').count(),0);
    assert.equal(await page.locator('.chapter-search select,.chapter-search .search-field svg').count(),0);
    const form=page.locator('.chapter-search');await form.getByRole('searchbox').fill('1');
    await form.locator('button').nth(1).click();await page.waitForURL(url=>url.searchParams.get('sort')==='desc');assert.equal(new URL(page.url()).searchParams.get('q'),'1');
    await page.locator('.chapter-search button').first().click();await page.waitForLoadState('networkidle');assert.equal(new URL(page.url()).searchParams.get('sort'),'desc');
    await page.locator('.chapter-search button').nth(1).click();await page.waitForURL(url=>url.searchParams.get('sort')==='asc');assert.equal(new URL(page.url()).searchParams.get('q'),'1');
   }finally{await context.close();}
  }
 }finally{await browser.close();}
});
