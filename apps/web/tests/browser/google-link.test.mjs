import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local fixture required');
test('settings Google link control uses rendered account, CSRF and private status',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});
 try{
 const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);const actor=(await login.json()).user;
 await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);
 const requests=[];
 await context.route('**/api/v1/auth/google/link',async route=>{requests.push(route.request());await route.fulfill({json:{enabled:true,linked:false}});});
 await context.route('**/api/v1/auth/google/link/start',async route=>{requests.push(route.request());await route.fulfill({status:409,json:{error:{code:'account_changed'}}});});
 const page=await context.newPage();await page.goto('/settings');
 const link=page.getByRole('button',{name:'Привязать Google'});await link.waitFor();
 assert.equal(await link.isEnabled(),true);await link.click();await page.getByRole('alert').waitFor();
 assert.equal(requests.length,2);assert.ok(requests.every(request=>request.headers()['x-qmanga-account']===actor.id));assert.equal(requests[1].method(),'POST');assert.ok(requests[1].headers()['x-csrf-token']);
 assert.equal(await page.locator('.profile-form input').inputValue(),actor.display_name);
 await context.unroute('**/api/v1/auth/google/link');
 await context.route('**/api/v1/auth/google/link',route=>route.fulfill({json:{enabled:true,linked:true}}));
 await page.goto('/settings?google_link=success');await page.getByText('Google привязан. Теперь можно входить в этот аккаунт через Google.').waitFor();assert.equal(await page.getByRole('button',{name:'Привязать Google'}).count(),0);
 for(const theme of ['light','dark'])for(const width of [320,360,1280]){await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);await page.setViewportSize({width,height:800});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);}
 await context.addCookies([{name:'qmanga_locale',value:'kk',url:baseURL}]);await page.goto('/settings');await page.getByText('Google байланыстырылған. Осы аккаунтқа Google арқылы кіре аласың.').waitFor();
 }finally{await context.close();await browser.close();}
});
