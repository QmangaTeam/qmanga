import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('cover statuses follow real bookmark updates on home and catalogue and clear on logout',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
   const account=(await (await context.request.get('/api/v1/auth/me')).json()).user.id;
   const page=await context.newPage();page.setDefaultTimeout(7000);await page.goto('/');
   const href=await page.locator('.cover-shelf .cover-card').first().getAttribute('href');assert.ok(href);
   const title=await (await context.request.get('/api/v1/titles/'+href.split('/').at(-1))).json();
   await page.goto(href);await page.locator('.library-trigger').click();const selection=page.locator('.action-dialog').filter({has:page.locator('.selection')});
   await selection.getByRole('button',{name:'Брошено',exact:true}).click();await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Брошено'));await selection.getByRole('button',{name:'Закрыть',exact:true}).click();
   const endpoint='/api/v1/reading/statuses?'+new URLSearchParams({ids:title.id+',absent'});
   const response=await context.request.get(endpoint,{headers:{'X-Qmanga-Account':account}});assert.equal(response.status(),200);assert.equal(response.headers()['cache-control'],'private, no-store');assert.deepEqual((await response.json()).statuses,{[title.id]:'dropped'});
   assert.equal((await context.request.get(endpoint,{headers:{'X-Qmanga-Account':'wrong'}})).status(),409);
   assert.equal((await context.request.get('/api/v1/reading/statuses?'+new URLSearchParams({ids:Array(61).fill(title.id).join(',')}),{headers:{'X-Qmanga-Account':account}})).status(),422);
   for(const path of ['/', '/catalog?q='+encodeURIComponent(title.name)]){
    await page.goto(path);const badge=page.locator(`.cover-card[href="${href}"] .reading-status`).first();await badge.waitFor();assert.equal(await badge.textContent(),'Брошено');
    for(const theme of ['dark','light'])for(const width of [320,360,1280]){
     await page.evaluate(value=>document.documentElement.dataset.theme=value,theme);await page.setViewportSize({width,height:850});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
     const position=await badge.evaluate(el=>{const b=el.getBoundingClientRect(),c=el.closest('.cover').getBoundingClientRect();return {right:c.right-b.right,top:b.top-c.top};});assert.ok(position.right<=8&&position.top<=8);
     if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/${path==='/'?'home':'catalog'}-${theme}-${width}.png`});}
    }
   }
   await page.goto(href);await page.locator('.library-trigger').click();await selection.getByRole('button',{name:'Читаю',exact:true}).click();await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Читаю'));
   await page.goto('/');await page.locator(`.cover-card[href="${href}"] .reading-status`).first().getByText('Читаю',{exact:true}).waitFor();
   await context.addCookies([{name:'qmanga_locale',value:'kk',url:fixture.app_url}]);await page.reload();await page.locator(`.cover-card[href="${href}"] .reading-status`).first().getByText('Оқып жүрмін',{exact:true}).waitFor();
   await page.getByRole('button',{name:'Аккаунт мәзірі',exact:true}).click();await page.locator('.account-menu').getByRole('button',{name:'Шығу',exact:true}).click();await page.waitForFunction(()=>document.querySelectorAll('.reading-status').length===0);
   assert.equal((await context.request.get(endpoint,{headers:{'X-Qmanga-Account':account}})).status(),401);
   await page.reload();assert.equal(await page.locator('.reading-status').count(),0);
  }finally{await context.close();}
 }finally{await browser.close();}
});
