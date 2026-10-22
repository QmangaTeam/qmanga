import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
const f=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(f.app_url).hostname));
test('title list picker stays open on failed write and closes only after a successful status save',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:f.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:f.app_url}]);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:f.app_url},data:{email:f.owner_email,password:f.owner_password}});assert.equal(login.status(),200);
   const account=(await login.json()).user.id;
   const response=await context.request.get('/api/v1/library/'+f.title_id,{headers:{'X-Qmanga-Account':account}});assert.equal(response.status(),200);
   let entry=(await response.json()).entry||{title_id:f.title_id,name:'Picker fixture',slug:f.title_slug,status:'reading',favorite:false,preferred_team_id:'',notify:true,version:0,progress:{}};
   let fail=true,releaseWrite,writeStarted;const started=new Promise(resolve=>writeStarted=resolve);
   const page=await context.newPage();page.setDefaultTimeout(6000);
   // Intercept writes: regression never changes the owner's persistent library.
   await page.route('**/api/v1/library/'+f.title_id,async route=>{
    if(route.request().method()==='PUT'){
     if(fail)return route.fulfill({status:503,json:{error:{code:'unavailable'}}});
     writeStarted();await new Promise(resolve=>releaseWrite=resolve);
     entry={...entry,...route.request().postDataJSON(),version:entry.version+1};return route.fulfill({status:204});
    }
    return route.fulfill({json:{entry}});
   });
   await page.goto('/titles/'+(f.title_slug||f.title_id));await page.locator('.library-trigger').click();
   const dialog=page.locator('dialog').filter({has:page.locator('.selection')});
   await dialog.getByRole('button',{name:'В планах',exact:true}).click();
   await dialog.getByText('Ещё не отправлены в облако.',{exact:false}).waitFor();assert.equal(await dialog.isVisible(),true);
   fail=false;await dialog.getByRole('button',{name:'Прочитано',exact:true}).click();await started;
   assert.equal(await dialog.isVisible(),true,'must stay open while write is pending');
   releaseWrite();await dialog.waitFor({state:'hidden'});
   assert.equal(await page.locator('.library-trigger').evaluate(e=>e===document.activeElement),true);
   assert.ok((await page.locator('.library-trigger').textContent()).includes('Прочитано'));
   await page.locator('.library-trigger').click();assert.equal(await dialog.getByRole('button',{name:'Прочитано',exact:true}).getAttribute('aria-pressed'),'true');
   await page.keyboard.press('Escape');
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:process.env.QMANGA_SCREENSHOT_DIR+'/library-picker-saved.png'});}
  }finally{await context.close();}
 }finally{await browser.close();}
});
