import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
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
