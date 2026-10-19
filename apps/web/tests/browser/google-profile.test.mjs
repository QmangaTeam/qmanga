import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local fixture required');
test('Google registration starts without manual nickname, email or password',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL});
 try{
 await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);
 await context.route('**/api/v1/auth/google/status',route=>route.fulfill({json:{enabled:true}}));
 let input;
 await context.route('**/api/v1/auth/google/start',async route=>{input=route.request().postDataJSON();await route.fulfill({status:422,json:{error:{code:'validation'}}});});
 const page=await context.newPage();await page.goto('/register?returnTo=%2Fcatalog');
 const googleButton=page.getByRole('button',{name:'Продолжить с Google'});
 assert.equal(await googleButton.locator('svg[aria-hidden="true"] path[fill="#4285F4"]').count(),1);
 await googleButton.click();
 await page.getByRole('alert').waitFor();
 assert.deepEqual(input,{return_path:'/catalog',nickname:''});
 assert.equal(await page.locator('#auth-nickname').inputValue(),'');
 }finally{await context.close();await browser.close();}
});
