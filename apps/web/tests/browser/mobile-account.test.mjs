import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
const baseURL=fixture.app_url;
assert.ok(['localhost','127.0.0.1'].includes(new URL(baseURL).hostname));
test('mobile Account consolidates navigation, role workspaces and recoverable team loading',{timeout:90000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const errors=[];
 try{
  for(const role of ['guest','owner','reader','admin']){
   const context=await browser.newContext({baseURL,viewport:{width:360,height:800},reducedMotion:'reduce'});
   try{
    if(role!=='guest'){
     const r=await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email:fixture[(role==='reader'&&!fixture.reader_email?'owner':role)+'_email'],password:fixture[(role==='reader'&&!fixture.reader_email?'owner':role)+'_password']}});assert.equal(r.status(),200);
    }
    if(role==='reader')await context.route('**/api/v1/studio/teams?*',route=>route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({items:[]})}));
    const page=await context.newPage();page.setDefaultTimeout(8000);page.on('pageerror',e=>errors.push(e.message));
    for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
     await context.addCookies([{name:'qmanga_locale',value:locale,url:baseURL}]);
     await page.goto('/menu');await page.evaluate(t=>document.documentElement.dataset.theme=t,theme);
     await page.getByRole('heading',{name:'Аккаунт',exact:true}).waitFor();
     if(role==='owner')await page.locator('.menu-page a[href="/studio/projects"]').waitFor();
     assert.equal(await page.locator('.menu-page a[href="/admin"]').count(),role==='admin'?1:0);
     if(role==='guest'||role==='reader')assert.equal(await page.locator('.menu-page a[href^="/studio"]').count(),0);
     assert.equal(await page.locator('.menu-page a[href="/catalog"],.menu-page a[href="/library"],.menu-page a[href="/notifications"]').count(),0);
     assert.equal(await page.locator('footer').count(),0);
     for(const width of [320,360,1280]){
      await page.setViewportSize({width,height:800});
      assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));
      const account=page.locator('.account-navigation');assert.equal(await account.isVisible(),width>=768);
      const nav=page.getByRole('navigation',{name:locale==='ru'?'Мобильная навигация':'Мобильді навигация',exact:true});
      assert.equal(await nav.isVisible(),width<768);
      if(width<768){assert.equal(await nav.locator('a').count(),5);assert.equal(await nav.getByRole('link',{name:'Аккаунт',exact:true}).getAttribute('aria-current'),'page');}
      if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/account-${role}-${locale}-${theme}-${width}.png`});}
     }
    }
    if(role==='owner'){
     let fail=true;
     await page.route('**/api/v1/studio/teams?*',route=>route.fulfill({status:fail?503:200,contentType:'application/json',body:JSON.stringify(fail?{error:{code:'unavailable'}}:{items:[]})}));
     await page.goto('/menu');await page.locator('.menu-page [role="alert"]').waitFor();assert.equal(await page.locator('.menu-page a[href="/studio/projects"]').count(),0);
     fail=false;await page.locator('.menu-page').getByRole('button',{name:'Повторить',exact:true}).click();await page.locator('.menu-page [role="alert"]').waitFor({state:'detached'});await page.locator('.menu-page [role="status"]').waitFor({state:'detached'});assert.equal(await page.locator('.menu-page a[href="/studio/projects"]').count(),0);
    }
    if(role==='admin'){
     await page.setViewportSize({width:1280,height:800});await page.getByRole('button',{name:'Меню аккаунта',exact:true}).click();await page.locator('dialog.account-menu[open]').waitFor();await page.keyboard.press('Escape');
     await page.setViewportSize({width:360,height:800});await page.locator('.menu-page').getByRole('button',{name:'Выйти',exact:true}).click();await page.waitForURL(baseURL+'/');await page.goto('/menu');await page.locator('.menu-page a[href^="/login"]').waitFor();assert.equal(await page.locator('.menu-page a[href="/admin"]').count(),0);
    }
   }finally{await context.close();}
  }
  assert.deepEqual(errors,[]);
 }finally{await browser.close();}
});
