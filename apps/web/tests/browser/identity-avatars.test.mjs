import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {join} from 'node:path';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local fixture required');
let browser;before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});after(async()=>await browser?.close());
test('nickname registration and sanitized avatars use the Qmanga origin',{timeout:90000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});
 try{
  const email=`avatar-${randomUUID()}@test.local`,password='Avatar-test-password-123';const origin={Origin:baseURL};
  for(const data of [{email,password},{email,password,display_name:'Reader@example.com'}]){
   const invalid=await context.request.post('/api/v1/auth/register',{headers:origin,data});assert.equal(invalid.status(),422);assert.equal((await invalid.json()).error.code,'validation');
  }
  assert.equal((await context.request.post('/api/v1/auth/register',{headers:origin,data:{email,password,display_name:'Аватар оқырманы'}})).status(),202);
  const login=await context.request.post('/api/v1/auth/login',{headers:origin,data:{email,password}});assert.equal(login.status(),200);const user=(await login.json()).user;assert.equal(user.display_name,'Аватар оқырманы');
  const page=await context.newPage();await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);
  await page.goto('/settings');const png=await page.evaluate(()=>{const canvas=document.createElement('canvas');canvas.width=480;canvas.height=240;const ctx=canvas.getContext('2d');ctx.fillStyle='#ff873d';ctx.fillRect(0,0,480,240);return canvas.toDataURL('image/png').split(',')[1];});
  const imageRequests=[];page.on('request',request=>{if(request.resourceType()==='image')imageRequests.push(request.url());});
  await page.locator('.avatar-editor input[type=file]').setInputFiles({name:'avatar.png',mimeType:'image/png',buffer:Buffer.from(png,'base64')});await page.getByRole('status').filter({hasText:'Аватар сохранён.'}).waitFor();
  const avatarURL='/media/avatars/user/'+user.id;const result=await context.request.get(avatarURL);assert.equal(result.status(),200);assert.equal(result.headers()['content-type'],'image/jpeg');assert.ok(result.headers()['x-content-type-options'].split(',').every(value=>value.trim()==='nosniff'));assert.equal((await context.request.head(avatarURL)).status(),200);
  const range=await context.request.get(avatarURL,{headers:{Range:'bytes=0-9'}});assert.equal(range.status(),206);assert.equal((await range.body()).length,10);
  assert.equal((await context.request.get(avatarURL,{headers:{Range:'bytes=0-1,4-5'}})).status(),416);
  assert.equal((await context.request.get('/media/avatars/user/missing-object')).status(),404);
  for(const theme of ['light','dark'])for(const width of [320,360,1280]){await page.evaluate(value=>{localStorage.setItem('qmanga.theme',value);document.documentElement.dataset.theme=value;},theme);await page.setViewportSize({width,height:800});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);assert.equal(await page.locator('.avatar-editor img').evaluate(img=>img.naturalWidth),256);if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`avatar-${theme}-${width}.png`)});}}
  assert.ok(imageRequests.every(url=>new URL(url).origin===new URL(baseURL).origin));
  await context.clearCookies();await page.goto('/register');await page.locator('#auth-nickname').waitFor();assert.equal(await page.locator('#auth-nickname').getAttribute('minlength'),'2');assert.equal(await page.locator('a[href*="google"]').count(),0);
  const status=await context.request.get('/api/v1/auth/google/status');assert.equal(status.status(),200);assert.equal((await status.json()).enabled,false);
 }finally{await context.close();}
});
