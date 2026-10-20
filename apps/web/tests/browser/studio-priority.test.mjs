import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Disposable local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
async function session(locale='ru',theme='dark'){
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800},reducedMotion:'reduce'});
 await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await context.addInitScript(theme=>localStorage.setItem('qmanga.theme',theme),theme);
 const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
 return context;
}

test('mobile project starts with current work and keeps chapter choice and metadata secondary',{timeout:45000},async()=>{
 for(const locale of ['ru','kk'])for(const theme of ['dark','light']){
 const context=await session(locale,theme);
 try{
  const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id);
  await page.locator('.draft-panel h2').waitFor();
  const pageCount=await page.locator('#draft-pages ol>li').count();
  assert.match(await page.locator('.revision-option[aria-current="true"]').innerText(),new RegExp((locale==='ru'?'Страниц: ':'Беттер: ')+pageCount),'opened revision uses full detail count rather than summary pages');
  for(const width of [320,360]){
   await page.setViewportSize({width,height:800});
   const panel=await page.locator('.draft-panel').boundingBox(),choice=await page.locator('.chapter-sidebar').boundingBox(),metadata=await page.locator('.cover-panel').boundingBox();
   assert.ok(panel.y<550,'current chapter should be on the first screen');assert.ok(choice.y>=panel.y+panel.height-1,'chapter choice follows current work');assert.ok(metadata.y>panel.y,'metadata follows work');
   assert.equal(await page.locator('.draft-panel').evaluate(el=>!!(el.compareDocumentPosition(document.querySelector('.chapter-sidebar'))&Node.DOCUMENT_POSITION_FOLLOWING)),true,'DOM order matches mobile reading order');
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
  }
  const change=page.getByRole('button',{name:locale==='ru'?'Выбрать главу':'Тарауды таңдау',exact:true});await change.press('Enter');
  await page.locator('.chapter-choice form').waitFor({state:'visible'});assert.equal(await page.locator('.chapter-choice summary').evaluate(el=>el===document.activeElement),true,'choice receives keyboard focus');
  await page.setViewportSize({width:1280,height:900});for(const id of ['draft-pages','draft-details','draft-preview'])assert.equal(await page.locator('#'+id).isVisible(),true,'desktop keeps all sections');
 }finally{await context.close();}
 }
});

test('project without chapters keeps creation reachable without a false current revision',{timeout:20000},async()=>{
 const context=await session();
 try{
  const page=await context.newPage();await page.route('**/api/v1/studio/projects/*/chapters?*',route=>route.fulfill({json:{items:[]}}));await page.route('**/api/v1/studio/projects/*/revisions?*',route=>route.fulfill({json:{items:[]}}));
  await page.goto('/studio/projects/'+fixture.project_id);await page.getByRole('textbox',{name:'Номер или обозначение главы'}).waitFor();
  assert.equal(await page.getByRole('button',{name:'Добавить главу и открыть черновик',exact:true}).isVisible(),true);assert.equal(await page.locator('.draft-panel h2').count(),0);
 }finally{await context.close();}
});


test('studio shell keeps site navigation and returns focus without reader chrome',{timeout:45000},async()=>{
 for(const locale of ['ru','kk'])for(const theme of ['dark','light']){
  const context=await session(locale,theme);
  try{
   const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id);await page.locator('.draft-panel h2').waitFor();
   const header=page.locator('.studio-header');assert.equal(await header.count(),1,'studio has its own compact header');
   assert.equal(await page.locator('.site-footer').count(),0);assert.equal(await page.getByRole('navigation',{name:locale==='ru'?'Мобильная навигация':'Мобильді навигация',exact:true}).count(),0);
   const trigger=header.getByRole('button',{name:locale==='ru'?'Навигация сайта':'Сайт навигациясы',exact:true});
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:800});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
    assert.ok((await trigger.boundingBox()).height>=44);await trigger.press('Enter');
    const dialog=page.getByRole('dialog',{name:locale==='ru'?'Навигация сайта':'Сайт навигациясы'});await dialog.waitFor();
    assert.deepEqual(await dialog.getByRole('link').evaluateAll(links=>links.map(link=>link.getAttribute('href'))),['/','/catalog','/library','/notifications','/menu']);
    await page.keyboard.press('Escape');assert.equal(await dialog.isVisible(),false);assert.equal(await trigger.evaluate(el=>el===document.activeElement),true);
   }
   await trigger.click();await page.getByRole('dialog').getByRole('link').filter({hasText:'Каталог'}).click();await page.waitForURL('**/catalog');
   assert.equal(await page.locator('.studio-header').count(),0);assert.equal(await page.locator('.site-footer').count(),1);
   await page.setViewportSize({width:360,height:800});assert.equal(await page.getByRole('navigation',{name:locale==='ru'?'Мобильная навигация':'Мобильді навигация',exact:true}).isVisible(),true);
  }finally{await context.close();}
 }
});


test('chapter groups retain paged revisions, mark current version and switch without creating',{timeout:60000},async()=>{
 for(const locale of ['ru','kk'])for(const theme of ['dark','light']){
 const context=await session(locale,theme);
 try{
  const response=await context.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions?offset=0');assert.equal(response.status(),200);
  const base=(await response.json()).items[0],chapterA=randomUUID(),chapterB=randomUUID(),chapterC=randomUUID();
  const revisions=Array.from({length:101},(_,index)=>({...base,id:randomUUID(),chapter_id:index<2?chapterA:index===100?chapterC:chapterB,chapter_label:index<2?'1':index===100?'Special':'2',state:index===0?'ready':'published'}));
  let creates=0,failSwitch=false;const page=await context.newPage();page.setDefaultTimeout(5000);
  await page.route('**/api/v1/studio/projects/*/chapters?*',route=>route.fulfill({json:{items:[{id:chapterA,label:'1',name:'First chapter',order:1},{id:chapterB,label:'2',name:'Second chapter',order:2}]}}));
  await page.route('**/api/v1/studio/projects/*/revisions?*',route=>route.fulfill({json:{items:new URL(route.request().url()).searchParams.get('offset')==='100'?revisions.slice(100):revisions.slice(0,100)}}));
  await page.route('**/api/v1/studio/projects/*/revisions',route=>{creates++;return route.fulfill({status:500,json:{error:{code:'internal'}}});});
  await page.route('**/api/v1/studio/revisions/*',route=>{if(failSwitch){failSwitch=false;return route.fulfill({status:500,json:{error:{code:'internal'}}});}const revision=revisions.find(item=>item.id===new URL(route.request().url()).pathname.split('/').at(-1));return revision?route.fulfill({json:revision}):route.continue();});
  await page.goto('/studio/projects/'+fixture.project_id);await page.locator('.draft-panel h2').waitFor();
  assert.equal(await page.locator('.chapter-group').count(),2,'logical chapters group all loaded versions');
  assert.equal(await page.locator('.revision-option').count(),100);assert.equal(await page.locator('.chapter-group').first().locator('.revision-option').count(),2);
  assert.match(await page.locator('.chapter-group').first().innerText(),/First chapter/);
  assert.equal(await page.locator('.revision-option[aria-current="true"]').count(),1);
  assert.match(await page.locator('.revision-option[aria-current="true"]').innerText(),locale==='ru'?/Открыта/:/Ашық/);
  const target=page.locator('.chapter-group').nth(1).locator('.revision-option').first();
  failSwitch=true;await target.press('Enter');await page.getByRole('alert').first().waitFor();
  assert.equal(await page.locator('.revision-option').first().getAttribute('aria-current'),'true','failed switch preserves current revision');
  assert.match(await page.locator('.draft-panel h2').innerText(),/1/);await target.press('Enter');
  await page.waitForFunction(()=>document.querySelector('.draft-panel h2')?.textContent.includes('2'));
  assert.equal(await target.getAttribute('aria-current'),'true');assert.equal(creates,0,'opening a revision is a read operation');
  await page.getByRole('button',{name:locale==='ru'?'Выбрать главу':'Тарауды таңдау',exact:true}).click();
  assert.equal(await page.locator('.chapter-choice select').nth(1).inputValue(),chapterB,'chapter selection matches opened revision');
  await page.getByRole('button',{name:locale==='ru'?'Ещё версии':'Тағы нұсқалар',exact:true}).click();
  await page.waitForFunction(()=>document.querySelectorAll('.revision-option').length===101);assert.equal(await page.locator('.chapter-group').count(),3,'revision without a loaded chapter remains reachable');
  await page.locator('.revision-option').first().click();await page.waitForFunction(()=>document.querySelector('.draft-panel h2')?.textContent.includes('1'));
  await page.locator('.draft-steps a[href="#draft-details"]').click();await page.locator('#draft-details textarea').first().fill('Unsaved source change');
  assert.equal(await target.isDisabled(),true,'dirty attribution blocks revision switching');assert.equal(creates,0);
  for(const width of [320,360,1280]){await page.setViewportSize({width,height:800});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);}
 }finally{await context.close();}
 }
});
