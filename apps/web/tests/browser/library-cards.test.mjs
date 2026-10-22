import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('library cover grid, inline filters and bookmark dialog fit phone and desktop',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
   const page=await context.newPage();page.setDefaultTimeout(7000);
   // Keep the real account selection; extra synthetic copies exercise a full grid without modifying catalogue data.
   const account=(await (await context.request.get('/api/v1/auth/me')).json()).user.id;
   const originalResponse=await context.request.get('/api/v1/library/'+fixture.title_id,{headers:{'X-Qmanga-Account':account}});assert.equal(originalResponse.status(),200);
   let original=await originalResponse.json();
   if(!original.entry||original.entry.removed){const csrf=(await(await context.request.get('/api/v1/auth/csrf')).json()).csrf_token;const saved=await context.request.put('/api/v1/library/'+fixture.title_id,{headers:{Origin:fixture.app_url,'X-CSRF-Token':csrf,'X-Qmanga-Account':account},data:{version:original.entry?.version||0,status:'reading',favorite:false,preferred_team_id:'',notify:true}});assert.equal(saved.status(),204);original=await(await context.request.get('/api/v1/library/'+fixture.title_id,{headers:{'X-Qmanga-Account':account}})).json();}
   assert.ok(original.entry,'fixture title must already belong to owner library');
   const entry=original.entry;assert.equal(typeof entry.cover_url,'string');
   const pattern='**/api/v1/library?*';
   await page.route(pattern,route=>route.fulfill({json:{items:Array.from({length:7},(_,i)=>({...entry,title_id:i?`grid-${i}`:entry.title_id,name:i?'Длинное название для проверки переноса '+i:entry.name,slug:i?`grid-${i}`:entry.slug})),has_more:false,next_cursor:'',total:7}}));
   await page.goto('/library');await page.locator('.library-card').first().waitFor();
   assert.equal(await page.locator('main input[type="search"],main .library-search,main a[href="/saved"]').count(),0);
   assert.equal(await page.locator('.library-filters button').count(),7);
   for(const theme of ['dark','light'])for(const width of [320,360,1280]){
    await page.evaluate(value=>document.documentElement.dataset.theme=value,theme);await page.setViewportSize({width,height:850});
    const boxes=await page.locator('.library-card').evaluateAll(items=>items.map(e=>{const r=e.getBoundingClientRect();return {x:r.x,y:r.y,width:r.width};}));
    const target=await page.locator('.library-card .options').first().boundingBox();assert.ok(target.width>=44&&target.height>=44);
    const count=boxes.filter(b=>Math.abs(b.y-boxes[0].y)<1).length;assert.equal(count,width<600?3:6);
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    if(entry.cover_url)assert.equal(await page.locator('.library-card .cover img').first().getAttribute('src'),entry.cover_url);
    await page.locator('.library-card .options').first().click();
    const dialog=page.locator('.library-card dialog[open]');await dialog.getByRole('combobox').waitFor();
    const bounds=await dialog.boundingBox();assert.ok(bounds.x>=0&&bounds.x+bounds.width<=width);
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/library-dialog-${theme}-${width}.png`});}
    await dialog.getByRole('button',{name:'Закрыть',exact:true}).click();
    assert.equal(await page.locator('.library-card .options').first().evaluate(e=>e===document.activeElement),true);
    if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/library-grid-${theme}-${width}.png`});
   }
   await page.unroute(pattern);await page.reload();await page.locator('.library-card .options').first().click();
   await page.locator('dialog[open] select').selectOption('planned');await page.getByText('Сохранено',{exact:true}).waitFor();
   await page.locator('dialog[open]').getByRole('button',{name:'Закрыть',exact:true}).click();
   await page.locator('.library-filters').getByRole('button',{name:'В планах',exact:true}).click();await page.locator('.library-card').first().waitFor();
   assert.equal((await (await context.request.get('/api/v1/library/'+fixture.title_id,{headers:{'X-Qmanga-Account':account}})).json()).entry.status,'planned');
   await page.setViewportSize({width:360,height:850});await page.locator('nav[aria-label="Мобильная навигация"] a[href="/menu"]').click();await page.locator('.menu-page a[href="/saved"]').waitFor();
  }finally{await context.close();}
 }finally{await browser.close();}
});
