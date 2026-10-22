import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Disposable fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
async function session(locale='ru',role='owner'){
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800},reducedMotion:'reduce'});await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);
 assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture[role+'_email'],password:fixture[role+'_password']}})).status(),200);return context;
}
test('team edit cancellation preserves fields and selected-team application link',{timeout:30000},async()=>{
 const context=await session();try{
 const page=await context.newPage();page.setDefaultTimeout(5000);await page.goto('/studio/teams');
 await page.getByRole('button',{name:'Изменить команду',exact:true}).click();const input=page.getByRole('textbox',{name:'Название команды',exact:true});await input.fill('Unsaved team name');
 const dismiss=page.waitForEvent('dialog').then(d=>d.dismiss());await Promise.all([dismiss,page.locator('.team-form').getByRole('button',{name:'Закрыть',exact:true}).click()]);assert.equal(await input.inputValue(),'Unsaved team name');
 const cancelLeave=page.waitForEvent('dialog').then(d=>d.dismiss());await Promise.all([cancelLeave,page.getByRole('link',{name:'Вернуться на сайт',exact:true}).click()]);assert.equal(new URL(page.url()).pathname,'/studio/teams');assert.equal(await input.inputValue(),'Unsaved team name');
 const link=page.getByRole('link',{name:'Запросить проект',exact:true});const team=(await(await context.request.get('/api/v1/studio/teams')).json()).items[0].id;assert.equal(new URL(await link.getAttribute('href'),fixture.app_url).searchParams.get('team'),team);
 const accept=page.waitForEvent('dialog').then(d=>d.accept());await Promise.all([accept,page.locator('.team-form').getByRole('button',{name:'Закрыть',exact:true}).click()]);assert.equal(await page.locator('.team-form').count(),0);
 }finally{await context.close();}
});
test('request resolves selected team beyond first page and inline errors preserve draft on failure',{timeout:45000},async()=>{
 const context=await session();try{
 const project=await(await context.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();const page=await context.newPage();page.setDefaultTimeout(5000);
 await page.route('**/api/v1/studio/teams?*',route=>{const offset=new URL(route.request().url()).searchParams.get('offset');return route.fulfill({json:{items:offset==='100'?[{id:project.team_id,name:'Chosen team',slug:'chosen',description:'',approved:true,my_role:'owner'}]:Array.from({length:100},(_,i)=>({id:'other-'+i,name:'Other '+i,slug:'other-'+i,description:'',approved:false,my_role:'member'}))}});});
 await page.goto('/studio/requests?new=1&team='+project.team_id);const form=page.locator('.request-form form');await form.waitFor();assert.equal(await form.getByRole('combobox',{name:'Команда',exact:true}).inputValue(),project.team_id);
 await form.getByRole('combobox',{name:'Что хочешь переводить?',exact:true}).selectOption('new');const note=form.getByRole('textbox',{name:'Планы команды',exact:true});await note.fill('Synthetic plan retained after validation and network failure');
 await form.getByRole('button',{name:'Отправить на рассмотрение',exact:true}).click();const summary=form.locator('.validation-summary');await summary.waitFor({state:'visible'});assert.equal(await summary.evaluate(el=>el===document.activeElement),true);
 const name=form.getByRole('textbox',{name:'Название на казахском',exact:true});assert.equal(await name.getAttribute('aria-invalid'),'true');assert.ok(await name.getAttribute('aria-describedby'));await summary.getByRole('button').first().click();assert.equal(await name.evaluate(el=>el===document.activeElement),true);await name.fill('Synthetic new title');assert.equal(await name.getAttribute('aria-invalid'),null);
 await page.route('**/api/v1/studio/project-requests',route=>route.request().method()==='POST'?route.fulfill({status:500,json:{error:{code:'internal'}}}):route.continue());
 await form.getByRole('button',{name:'Отправить на рассмотрение',exact:true}).click();await page.getByRole('alert').filter({hasNot:page.locator('ul')}).first().waitFor();assert.equal(await note.inputValue(),'Synthetic plan retained after validation and network failure');assert.equal(await name.inputValue(),'Synthetic new title');
 const cancel=page.waitForEvent('dialog').then(d=>d.dismiss());await Promise.all([cancel,page.locator('.request-form').getByRole('button',{name:'Закрыть',exact:true}).click()]);assert.equal(await name.inputValue(),'Synthetic new title');
 for(const width of [320,360,1280]){await page.setViewportSize({width,height:800});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);}
 const discard=page.waitForEvent('dialog').then(d=>d.accept());await Promise.all([discard,page.locator('.request-form').getByRole('button',{name:'Закрыть',exact:true}).click()]);
 await page.goto('/studio/requests?new=1&team=unavailable');await page.locator('.request-form form').waitFor();assert.equal(await page.locator('.request-form select').first().inputValue(),'');await page.getByText(/Для выбранной команды нельзя подать заявку/).waitFor();
 }finally{await context.close();}
});

test('administrative decision draft survives switching and navigation cancellation',{timeout:30000},async()=>{
 const context=await session('ru','admin');try{
 const page=await context.newPage();page.setDefaultTimeout(5000);await page.setViewportSize({width:1280,height:900});
 await page.route('**/api/v1/admin/teams/pending?*',route=>route.fulfill({json:{items:[1,2].map(n=>({id:'pending-'+n,name:'Pending '+n,slug:'pending-'+n,description:'Synthetic pending team',approved:false}))}}));
 await page.goto('/admin/teams');await page.getByRole('button',{name:'Вернуть на исправление',exact:true}).first().click();const reason=page.getByRole('textbox',{name:'Что нужно исправить?',exact:true});await reason.fill('Unsaved moderation reason');
 let cancel=page.waitForEvent('dialog').then(d=>d.dismiss());await Promise.all([cancel,page.getByRole('button',{name:'Вернуть на исправление',exact:true}).nth(1).click()]);assert.equal(await reason.inputValue(),'Unsaved moderation reason');assert.equal(await page.locator('article').first().locator('textarea').count(),1);
 cancel=page.waitForEvent('dialog').then(d=>d.dismiss());await Promise.all([cancel,page.locator('a[href="/admin"]').first().click()]);assert.equal(new URL(page.url()).pathname,'/admin/teams');assert.equal(await reason.inputValue(),'Unsaved moderation reason');
 }finally{await context.close();}
});
