import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('mobile title reflows and the single-Q wordmark follows explicit and system themes',{timeout:90000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const locale of ['kk','ru'])for(const mode of ['dark','light','system-dark','system-light']){
  const light=mode.endsWith('light');const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:900},colorScheme:light?'light':'dark',reducedMotion:'reduce'});
  try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);
   if(!mode.startsWith('system-'))await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),mode);
   const page=await context.newPage();await page.goto('/titles/'+fixture.title_slug);
   for(const width of [320,360,487]){
    await page.setViewportSize({width,height:900});
    const cover=await page.locator('.cover-wrap').boundingBox(),heading=await page.locator('.title-heading').boundingBox();
    assert.ok(cover.width>=170,'cover has useful phone size');assert.ok(heading.y+heading.height<=cover.y+1,'heading precedes cover');
    assert.equal(await page.locator('.title-information').evaluate(e=>e.scrollWidth<=e.clientWidth+1),true,'horizontal metadata stays contained in the panel');
    if(width===320){
     const rail=page.locator('.metadata');assert.equal(await rail.evaluate(e=>e.scrollWidth>e.clientWidth),true,'mobile metadata slides horizontally');
     await rail.focus();await page.keyboard.press('ArrowRight');await page.waitForFunction(()=>document.querySelector('.metadata').scrollLeft>0);
    }
    assert.equal(await page.locator('.title-tabs').evaluate(e=>e.scrollWidth<=e.clientWidth+1),true,'all title tabs fit');
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'page reflows');
    const actions=await page.locator('.title-actions').boundingBox(),read=await page.locator('.read-button').boundingBox();assert.ok(read.width>=actions.width-1,'Read spans phone width');
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:process.env.QMANGA_SCREENSHOT_DIR+`/title-${locale}-${mode}-${width}.png`,fullPage:true});}
   }
   for(const width of [768,885,960,1080,1440]){
    await page.setViewportSize({width,height:900});
    assert.equal(await page.locator('.metadata').count(),1,'single combined information panel');
    const info=await page.locator('.title-information').boundingBox(),tabs=await page.locator('.title-tabs').boundingBox(),cover=await page.locator('.cover-wrap').boundingBox();
    if(width<960)assert.ok(info.y>=tabs.y+tabs.height-1,'information follows tabs on smaller screens');
    else assert.ok(info.x+info.width<=tabs.x+1&&info.y>cover.y,'information sits below desktop cover');
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'information reflows at breakpoint');
   }
   await page.goto('/titles/'+fixture.title_slug+'?tab=about');await page.waitForLoadState('networkidle');
   for(const width of [320,688,1440]){await page.setViewportSize({width,height:900});
   const disclosure=page.locator('.description-toggle');assert.equal(await disclosure.getAttribute('aria-expanded'),'false');
   assert.equal(await page.locator('#title-description').evaluate(e=>getComputedStyle(e).webkitLineClamp),'3');
   await disclosure.click();await page.waitForFunction(()=>document.querySelector('.description-toggle')?.getAttribute('aria-expanded')==='true');assert.equal(await disclosure.getAttribute('aria-expanded'),'true');assert.equal(await page.locator('#title-description').evaluate(e=>getComputedStyle(e).webkitLineClamp),'none');
   assert.equal(await disclosure.innerText(),locale==='kk'?'Жию':'Свернуть');await disclosure.click();await page.waitForFunction(()=>document.querySelector('.description-toggle')?.getAttribute('aria-expanded')==='false');assert.equal(await disclosure.getAttribute('aria-expanded'),'false');
   }
   await page.locator('.statistics').click();assert.equal(await page.locator('.stats-dialog').evaluate(e=>e.open),true);await page.keyboard.press('Escape');
   await page.setViewportSize({width:1280,height:900});const mark=page.locator('header a.wordmark');
   assert.equal(await mark.locator('img').count(),1);assert.equal(await mark.locator('.brand-name').innerText(),'manga.');assert.equal(await mark.locator('img').getAttribute('src'),'/favicon.png');
   const color=await mark.locator('.brand-name').evaluate(e=>getComputedStyle(e).color);assert.equal(color,light?'rgb(33, 27, 24)':'rgb(239, 237, 241)');
   if(process.env.QMANGA_SCREENSHOT_DIR){await page.evaluate(()=>scrollTo(0,0));await page.screenshot({path:process.env.QMANGA_SCREENSHOT_DIR+`/desktop-${locale}-${mode}.png`,fullPage:true});}
  }finally{await context.close();}
 }}finally{await browser.close();}
});
