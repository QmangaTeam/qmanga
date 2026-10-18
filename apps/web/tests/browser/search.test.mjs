import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {mkdir} from 'node:fs/promises';
import {join} from 'node:path';
import {readFile} from 'node:fs/promises';
const fixture=process.env.QMANGA_FIXTURE_FILE?JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')):null;
const baseURL=fixture?.app_url||process.env.QMANGA_TEST_URL||'http://localhost:8080';
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local-only search regression');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
for(const locale of ['kk','ru'])for(const theme of ['light','dark']){
 test(`search has one focus outline and returns keyboard focus: ${locale}/${theme}`,{timeout:30000},async()=>{
  const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:baseURL}]);
   await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   const page=await context.newPage();page.setDefaultTimeout(5000);page.setDefaultNavigationTimeout(10000);await page.goto('/');
   const trigger=page.locator('.search-trigger');
   await trigger.click();
   const dialog=page.getByRole('dialog');const input=dialog.getByRole('searchbox');
   await input.fill('Local & test');
   const styles=await input.evaluate(el=>({focused:el===document.activeElement,inputOutline:getComputedStyle(el).outlineStyle,groupOutline:getComputedStyle(el.parentElement).outlineStyle,groupWidth:getComputedStyle(el.parentElement).outlineWidth}));
   assert.equal(styles.focused,true);
   assert.equal(styles.inputOutline,'none','embedded input must not paint a second outline');
   assert.equal(styles.groupOutline,'solid','the complete search field must show focus');
   assert.equal(styles.groupWidth,'2px');
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:800});
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
    const bounds=await dialog.boundingBox();assert.ok(bounds.x>=0&&bounds.x+bounds.width<=width+1);
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`search-${locale}-${theme}-${width}.png`)});}
   }
   await input.press('Tab');
   assert.equal(await dialog.getByRole('button',{name:locale==='ru'?'Найти':'Іздеу',exact:true}).evaluate(el=>el===document.activeElement),true);
   await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
   assert.equal(await trigger.evaluate(el=>el===document.activeElement),true,'Escape restores focus to search opener');
   await trigger.click();await dialog.getByRole('button',{name:locale==='ru'?'Закрыть':'Жабу',exact:true}).click();
   assert.equal(await trigger.evaluate(el=>el===document.activeElement),true);
   await trigger.click();await input.fill('Local & test');await input.press('Enter');
   await page.waitForURL(url=>url.pathname==='/catalog'&&url.searchParams.get('q')==='Local & test');
   assert.equal(await page.getByRole('dialog').isVisible(),false);
  }finally{await context.close();}
 });
}
