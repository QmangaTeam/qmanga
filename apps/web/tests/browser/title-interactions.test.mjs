import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {join} from 'node:path';
import {chromium} from 'playwright';

if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Disposable local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});

async function assertNoHorizontalOverflow(page){
 const measure=await page.evaluate(()=>({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,offenders:[...document.querySelectorAll('body *')].map(element=>({tag:element.tagName,className:typeof element.className==='string'?element.className:'',text:element.textContent?.trim().slice(0,35),left:Math.round(element.getBoundingClientRect().left),right:Math.round(element.getBoundingClientRect().right)})).filter(item=>item.right>innerWidth+1||item.left< -1).slice(0,12)}));
 assert.equal(measure.scrollWidth<=measure.width+1,true,JSON.stringify(measure));
}

test('title stats live inside facts and comments load without an extra click', {timeout:30000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:320,height:750}});
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const title={id:fixture.title_id,slug:fixture.title_slug};
  assert.ok(title.slug);
  const detail=(await (await context.request.get('/api/v1/titles/'+title.slug)).json());
  const page=await context.newPage();
  await page.goto('/titles/'+title.slug+'?tab=about');
  await page.waitForLoadState('networkidle');
  assert.equal(await page.locator('.title-tabs a').count(),3);
  assert.equal(await page.locator('.metadata a[href*="original_status=ongoing"] .status-dot.ongoing').count(),1);
  await page.getByRole('button',{name:'Статистика тайтла',exact:true}).click();
  const dialog=page.getByRole('dialog',{name:'Статистика тайтла'});
  await dialog.waitFor();
  if(detail.rating_count===0)assert.equal(await dialog.getByText('Оценок пока нет').count(),1);
  else assert.equal(await dialog.locator('.rating-bar').count(),10);
  if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'title-stats-320.png')});}
  await dialog.getByRole('button',{name:'Закрыть'}).click();
  await assertNoHorizontalOverflow(page);
  const request=page.waitForResponse(response=>response.url().includes('/api/v1/community/titles/')&&response.request().method()==='GET');
  await page.getByRole('link',{name:'Комментарии'}).click();
  assert.equal((await request).status(),200);
  await page.locator('.discussion h2').waitFor();
  assert.equal(await page.getByRole('button',{name:'Открыть комментарии'}).count(),0);
 }finally{await context.close();}
});

test('title rating saves independently and updates its distribution', {timeout:30000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:1280,height:800}});
 let original=0,headers={};
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}});
  assert.equal(login.status(),200);
  const account=(await login.json()).user;
  const csrf=(await context.cookies(fixture.app_url)).find(cookie=>cookie.name==='qmanga_csrf')?.value;
  headers={Origin:fixture.app_url,'X-Qmanga-Account':account.id,'X-CSRF-Token':csrf};
  const title={id:fixture.title_id,slug:fixture.title_slug};
  assert.ok(title.slug);
  const current=await context.request.get('/api/v1/community/titles/'+fixture.title_id+'/rating',{headers});
  assert.equal(current.status(),200);
  original=(await current.json()).score;
  const page=await context.newPage();
  await page.goto('/titles/'+title.slug+'?tab=about');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button',{name:'Оценить тайтл'}).click();
  const dialog=page.getByRole('dialog',{name:'Оценить тайтл'});
  await dialog.getByRole('button',{name:'7 / 10',exact:true}).click();
  await dialog.waitFor({state:'hidden'});
  await page.getByRole('button',{name:'Статистика тайтла',exact:true}).click();
  const bar=page.locator('.stats-dialog .rating-bar').filter({has:page.locator('span').filter({hasText:/^7$/})});
  assert.equal(await bar.count()>0,true);
  const detail=await (await context.request.get('/api/v1/titles/'+title.slug)).json();
  assert.ok(detail.rating_distribution[6]>=1);
  await page.goto('/titles/'+title.slug+'?tab=comments');
  await page.locator('.composer-prompt').click();
  await page.getByRole('textbox',{name:'Написать комментарий'}).waitFor();
  assert.equal(await page.locator('.format-actions button').count(),4);
  if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'comments-editor-1280.png')});}
  await page.setViewportSize({width:320,height:750});
  await assertNoHorizontalOverflow(page);
  if(process.env.QMANGA_SCREENSHOT_DIR){await page.locator('.discussion').scrollIntoViewIfNeeded();await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'comments-editor-320.png')});}
 }finally{
  await context.request.put('/api/v1/community/titles/'+fixture.title_id+'/rating',{headers,data:{score:original}}).catch(()=>{});
  await context.close();
 }
});

// Reading regressions own the owner account's library; use the separate verified admin here.
test('title actions open dialogs and library status saves with one choice', {timeout:30000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:780}});
 let headers={},original=null,changed=null;
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}});
  assert.equal(login.status(),200);
  const account=(await login.json()).user;
  const csrf=(await context.cookies(fixture.app_url)).find(cookie=>cookie.name==='qmanga_csrf')?.value;
  headers={Origin:fixture.app_url,'X-Qmanga-Account':account.id,'X-CSRF-Token':csrf};
  const title={id:fixture.title_id,slug:fixture.title_slug};
  assert.ok(title.slug);
  original=(await (await context.request.get('/api/v1/library/'+title.id,{headers})).json()).entry;
  const page=await context.newPage();
  await page.goto('/titles/'+title.slug);
  await page.waitForLoadState('networkidle');
  await page.getByRole('button',{name:'Добавить в список'}).click();
  const dialog=page.getByRole('dialog',{name:'Мой список'});
  await dialog.getByRole('button',{name:'В планах'}).click();
  await dialog.getByRole('button',{name:'В планах'}).waitFor({state:'visible'});
  changed=(await (await context.request.get('/api/v1/library/'+title.id,{headers})).json()).entry;
  assert.equal(changed.status,'planned');
  await dialog.getByRole('button',{name:'Закрыть'}).click();
  await page.getByRole('button',{name:'Сообщить о проблеме'}).click();
  const report=page.getByRole('dialog',{name:'Сообщить о проблеме'});
  await report.getByRole('button',{name:'Отправить жалобу'}).waitFor();
  assert.equal(await report.locator('summary').count(),0);
  assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
 }finally{
  if(changed){
   const cleanup=original
    ? await context.request.put('/api/v1/library/'+fixture.title_id,{headers,data:{version:changed.version,removed:original.removed===true,status:original.status,favorite:original.favorite,preferred_team_id:original.preferred_team_id,notify:original.notify}})
    : await context.request.delete('/api/v1/library/'+fixture.title_id,{headers});
   assert.equal(cleanup.status(),204,'restore the fixture library state');
   const restored=await (await context.request.get('/api/v1/library/'+fixture.title_id,{headers})).json();
   assert.equal(restored.entry?.removed??true,original?.removed??true,'cleanup must not resurrect a removed entry');
  }
  await context.close();
 }
});
