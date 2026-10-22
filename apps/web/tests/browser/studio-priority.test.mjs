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

test('project overview offers internal sections and keeps editor behind explicit continue',{timeout:45000},async()=>{
 for(const locale of ['ru','kk'])for(const theme of ['dark','light']){const context=await session(locale,theme);try{const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id);await page.locator('.version-row a').first().waitFor();assert.equal(await page.locator('.draft-panel').count(),0);assert.equal(await page.locator('.project-navigation a').count(),4);for(const width of [320,360,1280]){await page.setViewportSize({width,height:900});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));}await page.locator('.version-row a').first().click();await page.locator('.draft-panel h2').waitFor();const change=page.getByRole('button',{name:locale==='ru'?'Выбрать главу':'Тарауды таңдау',exact:true});await change.press('Enter');assert.equal(await page.locator('.chapter-switch-dialog').evaluate(el=>el.matches(':modal')),true);await page.keyboard.press('Escape');assert.equal(await change.evaluate(el=>el===document.activeElement),true);for(const id of ['draft-pages','draft-details','draft-preview']){await page.locator('.draft-steps a[href="#'+id+'"]').click();assert.equal(await page.locator('#'+id).isVisible(),true);}}finally{await context.close();}}
});
test('empty project offers creation without a false current revision',{timeout:20000},async()=>{
 const context=await session();try{const page=await context.newPage();await page.route('**/api/v1/studio/projects/*/chapters?*',route=>route.fulfill({json:{items:[]}}));await page.route('**/api/v1/studio/projects/*/revisions?*',route=>route.fulfill({json:{items:[]}}));await page.goto('/studio/projects/'+fixture.project_id);await page.getByRole('button',{name:'Добавить главу',exact:true}).click();await page.getByRole('textbox',{name:'Номер или обозначение главы'}).waitFor();assert.equal(await page.locator('.draft-panel h2').count(),0);}finally{await context.close();}
});


test('studio shell keeps catalog access and responsive workspace navigation without reader chrome',{timeout:45000},async()=>{
 for(const locale of ['ru','kk'])for(const theme of ['dark','light']){
  const context=await session(locale,theme);
  try{const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id);await page.locator('.project-heading .identity p').waitFor();
   assert.equal(await page.locator('.workspace-topbar').count(),1);assert.equal(await page.locator('.site-footer').count(),0);
   for(const width of [320,360,1280]){await page.setViewportSize({width,height:800});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    if(width<1000){const trigger=page.locator('.workspace-menu');await trigger.press('Enter');const dialog=page.locator('.workspace-drawer');await dialog.waitFor();assert.equal(await dialog.locator('nav a').count(),4);await page.keyboard.press('Escape');assert.equal(await trigger.evaluate(el=>el===document.activeElement),true);}
   }
   await page.locator('.workspace-catalog').click();await page.waitForURL('**/catalog');assert.equal(await page.locator('.workspace-topbar').count(),0);assert.equal(await page.locator('.site-footer').count(),1);
  }finally{await context.close();}
 }
});
