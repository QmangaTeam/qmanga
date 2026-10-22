import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
async function seedGuest(page,manifest){await page.evaluate(async item=>{
 const connection=await new Promise((resolve,reject)=>{const request=indexedDB.open('qmanga-reading',2);request.onupgradeneeded=()=>{const store=request.result.createObjectStore('positions',{keyPath:['scope','titleId']});store.createIndex('recent',['scope','updatedAt','titleId']);};request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error);});
 await new Promise((resolve,reject)=>{const work=connection.transaction('positions','readwrite');work.objectStore('positions').put({scope:'guest',titleId:item.title_id,name:item.title_name,slug:item.title_slug,updatedAt:Date.now(),progress:{titleId:item.title_id,releaseId:item.release_id,revisionId:item.revision_id,pageId:item.pages[0].id,fraction:.5,version:0,furthestOrder:0,deviceId:'guest-device',sequence:1},pending:null,conflict:null});work.oncomplete=()=>{connection.close();resolve();};work.onerror=()=>reject(work.error);});
 },manifest);}
test('new registration imports automatically, retries a failed transfer and never repeats at ordinary login',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const address='reading-import-'+randomUUID()+'@test.local',password='Import-test-'+randomUUID();
 try{
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
   const response=await context.request.get('/api/v1/releases/'+fixture.release_id+'/manifest');assert.equal(response.status(),200);const manifest=await response.json();
   const page=await context.newPage();const errors=[];page.on('pageerror',error=>errors.push(error.message));page.setDefaultTimeout(8000);await page.goto('/');await seedGuest(page,manifest);
   await page.goto('/register');await page.waitForLoadState('networkidle');await page.locator('#auth-nickname').fill('Import Reader');await page.locator('#auth-email').fill(address);await page.locator('#auth-password').fill(password);await page.locator('.terms-consent input').check();await page.getByRole('button',{name:'Регистрация',exact:true}).click();await page.getByText('Письмо отправлено. Проверь почту.',{exact:true}).waitFor();
   const receipt=(await context.cookies()).find(cookie=>cookie.name==='qmanga_guest_registration');assert.ok(receipt?.httpOnly);assert.equal(receipt.sameSite,'Lax');
   await page.route('**/api/v1/progress/*',route=>route.request().method()==='POST'?route.fulfill({status:500,json:{error:{code:'internal'}}}):route.continue());
   await page.goto('/login?returnTo=%2Flibrary');await page.waitForLoadState('networkidle');await page.locator('#auth-email').fill(address);await page.locator('#auth-password').fill(password);await page.getByRole('button',{name:'Войти',exact:true}).click();await page.waitForURL('**/library');await page.locator('.import-notice').waitFor();
   const account=(await (await context.request.get('/api/v1/auth/me')).json()).user.id,headers={'X-Qmanga-Account':account};
   assert.equal((await (await context.request.get('/api/v1/auth/guest-import',{headers})).json()).allowed,true);
   await page.unroute('**/api/v1/progress/*');await page.evaluate(()=>window.dispatchEvent(new Event('focus')));
   await page.locator('.library-card').filter({has:page.locator(`.card-heading a[href="/titles/${manifest.title_slug}"]`)}).waitFor();
   await page.locator('.import-notice').waitFor({state:'hidden'});
   assert.equal(await page.getByRole('heading',{name:'Прочитано гостем',exact:true}).count(),0);
   const progress=await (await context.request.get('/api/v1/progress/'+manifest.title_id,{headers})).json();assert.equal(progress.release_id,manifest.release_id);assert.equal(progress.fraction,.5);
   assert.equal((await (await context.request.get('/api/v1/auth/guest-import',{headers})).json()).allowed,false);
   await page.getByRole('button',{name:'Меню аккаунта',exact:true}).click();await page.locator('.account-menu').getByRole('button',{name:'Выйти',exact:true}).click();await page.waitForURL(url=>url.pathname==='/');assert.deepEqual(errors,[]);
   await seedGuest(page,{...manifest});
   await page.goto('/login?returnTo=%2Flibrary');await page.waitForLoadState('networkidle');await page.locator('#auth-email').fill(address);await page.locator('#auth-password').fill(password);await page.getByRole('button',{name:'Войти',exact:true}).click();await page.waitForURL('**/library');
   const again=await (await context.request.get('/api/v1/progress/'+manifest.title_id,{headers})).json();assert.equal(again.version,progress.version,'ordinary login must not import again');
   // A valid receipt from this browser grants no access to an unrelated existing account.
   const old=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(old.status(),200);
   const oldAccount=(await old.json()).user.id;assert.equal((await (await context.request.get('/api/v1/auth/guest-import',{headers:{'X-Qmanga-Account':oldAccount}})).json()).allowed,false);
  }finally{await context.close();}
 }finally{await browser.close();}
});
