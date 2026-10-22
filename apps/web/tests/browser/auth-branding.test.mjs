import {test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {readFile} from 'node:fs/promises';
const fixture=process.env.QMANGA_FIXTURE_FILE?JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')):null;
const appURL=fixture?fixture.app_url:process.env.QMANGA_APP_URL;
assert.ok(appURL,'QMANGA_FIXTURE_FILE or explicit QMANGA_APP_URL required');
assert.ok(['localhost','127.0.0.1'].includes(new URL(appURL).hostname));
test('auth branding and centered mobile Home retain localized navigation and touch targets',{timeout:90000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
  const context=await browser.newContext({baseURL:appURL,reducedMotion:'reduce'});
  try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:appURL}]);
   await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   const page=await context.newPage();
   for(const width of [320,360,1280])for(const route of ['/login?returnTo=%2Flibrary','/register?returnTo=%2Flibrary']){
    await page.setViewportSize({width,height:900});await page.goto(route);
    await page.waitForFunction(value=>document.documentElement.dataset.theme===value,theme);
    const brand=page.locator('.auth-brand');await brand.waitFor();
    assert.equal(await brand.getAttribute('aria-label'),'Qmanga');assert.equal(await brand.locator('img').count(),1);
    assert.equal(await brand.locator('img').evaluate(e=>e.complete&&e.naturalWidth>0),true);
    assert.equal(await brand.locator('img').getAttribute('src'),'/favicon.png');
    assert.equal(await brand.innerText(),'');
    assert.equal(await brand.locator('img').getAttribute('width'),'32');
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
    assert.equal(await brand.evaluate(e=>getComputedStyle(e).color),theme==='light'?'rgb(33, 27, 24)':'rgb(239, 237, 241)');
   }
   await page.goto('/catalog');await page.locator('nav[aria-label="'+(locale==='kk'?'Мобильді навигация':'Мобильная навигация')+'"]').waitFor({state:'attached'});
   const nav=page.locator('nav[aria-label="'+(locale==='kk'?'Мобильді навигация':'Мобильная навигация')+'"]');
   for(const width of [320,360]){
    await page.setViewportSize({width,height:900});
    assert.deepEqual(await nav.locator('a').evaluateAll(links=>links.map(link=>link.getAttribute('href'))),['/catalog','/','/library','/notifications']);
    assert.equal(await nav.locator('a').nth(1).getAttribute('aria-label'),locale==='kk'?'Басты бет':'Главная');
    assert.equal(await nav.locator('a[href="/"] img').count(),1);assert.equal(await nav.locator('a[href="/"] svg').count(),0);
    for(const link of await nav.locator('a').all()){const box=await link.boundingBox();assert.ok(box.width>=44&&box.height>=44);}
    const home=nav.locator('a[href="/"]');await home.focus();assert.equal(await home.evaluate(e=>e===document.activeElement),true);
    const box=await home.boundingBox();assert.ok(Math.abs(box.x+box.width/2-width/2)<1);
   }
  }finally{await context.close();}
 }}finally{await browser.close();}
});
