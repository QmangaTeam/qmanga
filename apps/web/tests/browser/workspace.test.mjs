import {test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {mkdir} from 'node:fs/promises';
const baseURL=process.env.QMANGA_TEST_URL||'http://localhost:8080';
if(!['localhost','127.0.0.1','::1'].includes(new URL(baseURL).hostname))throw new Error('Local-only browser checks');
test('studio sidebar adapts to phone disclosure, supports keyboard and closes after navigation', {timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  for(const width of [320,360,1280])for(const theme of ['light','dark']){
   const context=await browser.newContext({baseURL,viewport:{width,height:800},reducedMotion:'reduce'});
   try{
    await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);
    await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
    const page=await context.newPage();await page.goto('/studio/projects');
    const nav=page.getByRole('navigation',{name:'Кабинет переводчика'});
    const toggle=page.getByRole('button',{name:/Разделы/});
    if(width<900){
     assert.equal(await nav.isVisible(),false);
     await toggle.focus();await page.keyboard.press('Enter');
     await nav.waitFor({state:'visible'});assert.equal(await toggle.getAttribute('aria-expanded'),'true');
     await page.keyboard.press('Tab');assert.equal(await nav.locator('a').first().evaluate(el=>el===document.activeElement),true);
    }else{assert.equal(await toggle.isVisible(),false);assert.equal(await nav.isVisible(),true);}
    assert.equal(await nav.locator('a').count(),4);
    assert.equal(await nav.locator('a[aria-current="page"]').getAttribute('href'),'/studio/projects');
    const navBox=await nav.boundingBox();const contentBox=await page.locator('.workspace-content').boundingBox();
    if(width>=900)assert.ok(navBox.x+navBox.width<=contentBox.x,'desktop navigation is beside content');
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'no page overflow');
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/workspace-${width}-${theme}.png`});}
    await nav.getByRole('link',{name:'Заявки',exact:true}).click();await page.waitForURL('**/studio/requests');
    if(width<900){await nav.waitFor({state:'hidden'});assert.equal(await toggle.getAttribute('aria-expanded'),'false');}
   }finally{await context.close();}
  }
 }finally{await browser.close();}
});
test('settings keep appearance after account in the visible page order',{timeout:15000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{const page=await browser.newPage({baseURL});await page.goto('/settings');const headings=await page.locator('.settings>section>h2').allTextContents();assert.deepEqual(headings,['Аккаунт','Сыртқы көрініс','Оқу']);}finally{await browser.close();}
});
