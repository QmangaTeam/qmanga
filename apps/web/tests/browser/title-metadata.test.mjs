import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('title counts match the API, facts stay in one row and icon sorting preserves chapter queries',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
   const context=await browser.newContext({baseURL:fixture.app_url});
   try{
    await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
    const data=await (await context.request.get('/api/v1/titles/'+(fixture.title_slug||fixture.title_id))).json();assert.equal(typeof data.comment_count,'number');
    const page=await context.newPage();page.setDefaultTimeout(7000);await page.goto('/titles/'+data.slug);
    assert.equal(await page.locator('.title-tabs a[href*="tab=chapters"] .tab-count').textContent(),String(data.chapter_count));
    assert.equal(await page.locator('.title-tabs a[href*="tab=comments"] .tab-count').textContent(),String(data.comment_count));
    for(const width of [320,360,1280]){
     await page.setViewportSize({width,height:850});
     const rects=await page.locator('.title-facts>div').evaluateAll(elements=>elements.map(e=>e.getBoundingClientRect().y));assert.ok(rects.every(y=>Math.abs(y-rects[0])<2),'all facts remain in one row');
     assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
     if(width===320){const facts=page.locator('.title-facts');assert.ok(await facts.evaluate(el=>el.scrollWidth>el.clientWidth),'narrow facts row scrolls internally');await facts.focus();await page.keyboard.press('ArrowRight');await page.waitForFunction(()=>document.querySelector('.title-facts').scrollLeft>0);await facts.evaluate(el=>el.scrollLeft=0);}
     if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/title-${locale}-${theme}-${width}.png`});}
    }
    assert.equal(await page.locator('.chapter-search select,.chapter-search .search-field svg').count(),0);
    const form=page.locator('.chapter-search');await form.getByRole('searchbox').fill('1');
    await form.locator('button').nth(1).click();await page.waitForURL(url=>url.searchParams.get('sort')==='desc');assert.equal(new URL(page.url()).searchParams.get('q'),'1');
    await page.locator('.chapter-search button').first().click();await page.waitForLoadState('networkidle');assert.equal(new URL(page.url()).searchParams.get('sort'),'desc');
    await page.locator('.chapter-search button').nth(1).click();await page.waitForURL(url=>url.searchParams.get('sort')==='asc');assert.equal(new URL(page.url()).searchParams.get('q'),'1');
   }finally{await context.close();}
  }
 }finally{await browser.close();}
});
