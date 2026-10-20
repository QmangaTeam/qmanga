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
    assert.equal(await page.locator('.title-facts').evaluate(e=>e.scrollWidth<=e.clientWidth+1),true,'facts never require side scrolling');
    assert.equal(await page.locator('.title-tabs').evaluate(e=>e.scrollWidth<=e.clientWidth+1),true,'all title tabs fit');
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'page reflows');
    const actions=await page.locator('.title-actions').boundingBox(),read=await page.locator('.read-button').boundingBox();assert.ok(read.width>=actions.width-1,'Read spans phone width');
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:process.env.QMANGA_SCREENSHOT_DIR+`/title-${locale}-${mode}-${width}.png`,fullPage:true});}
   }
   for(const width of [885,960,1280]){
    await page.setViewportSize({width,height:900});
    const cells=await page.locator('.title-facts>div').evaluateAll(items=>items.map(item=>({width:item.getBoundingClientRect().width,label:item.querySelector('dt').getBoundingClientRect().bottom,value:item.querySelector('dd').getBoundingClientRect().top})));
    for(const cell of cells){assert.ok(Math.abs(cell.value-cell.label-4)<1,'compact four-pixel label/value gap');assert.ok(Math.abs(cell.width-cells[0].width)<1,'equal fact column widths');assert.ok(Math.abs(cell.label-cells[0].label)<1,'aligned fact label baselines');assert.ok(Math.abs(cell.value-cells[0].value)<1,'aligned fact values');}
   }
   await page.setViewportSize({width:1280,height:900});const mark=page.locator('header a.wordmark');
   assert.equal(await mark.locator('img').count(),1);assert.equal(await mark.locator('.brand-name').innerText(),'manga.');assert.equal(await mark.locator('img').getAttribute('src'),'/favicon.png');
   const color=await mark.locator('.brand-name').evaluate(e=>getComputedStyle(e).color);assert.equal(color,light?'rgb(33, 27, 24)':'rgb(239, 237, 241)');
   if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:process.env.QMANGA_SCREENSHOT_DIR+`/desktop-${locale}-${mode}.png`});
  }finally{await context.close();}
 }}finally{await browser.close();}
});
