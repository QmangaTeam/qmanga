import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Local fixture only');
async function session(){const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);const result=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(result.status(),200);return {browser,context};}
test('preflight, processing summary and page overview follow actual revision state',{timeout:60000},async()=>{
 const {browser,context}=await session();try{
 const page=await context.newPage();let ready=false;
 const revision=()=>({id:'workflow-revision',release_id:fixture.release_id,project_id:fixture.project_id,chapter_id:'workflow-chapter',title_id:fixture.title_id,title_name:'Workflow title',chapter_label:'Workflow',team_name:'Synthetic team',state:ready?'ready':'processing',note:'',version:1,source_note:ready?'Licensed synthetic source':'',credits:ready?'Synthetic team':'',rights_confirmed:ready,pages:['ready',ready?'ready':'failed',ready?'ready':'queued'].map((status,index)=>({id:'workflow-page-'+index,position:index+1,width:600,height:900,bytes:100,status,failure:status==='failed'?'processing_failed':'',upload_key:'workflow-key-'+index}))});
 await page.route('**/api/v1/studio/projects/*/revisions?*',route=>route.fulfill({json:{items:[revision()]}}));
 await page.route('**/api/v1/studio/revisions/workflow-revision',route=>route.fulfill({json:revision()}));
 await page.goto('/studio/projects/'+fixture.project_id+'?revision=workflow-revision&step=draft-preview');await page.locator('.preflight').waitFor();
 assert.match(await page.locator('.preflight').innerText(),/Исправь страницы с ошибками/);assert.match(await page.locator('.preflight').innerText(),/источник/);assert.equal(await page.getByRole('button',{name:'Отправить на проверку',exact:true}).isDisabled(),true);
 await page.getByRole('button',{name:'Исправь страницы с ошибками',exact:true}).click();
 assert.match(await page.locator('.processing-summary').innerText(),/Готово: 1 · Обработка: 1 · Ошибки: 1/);
 await page.getByRole('button',{name:'К странице с ошибкой 2',exact:true}).click();assert.equal(await page.locator('#studio-page-workflow-page-1').evaluate(el=>el===document.activeElement),true);
 await page.getByRole('button',{name:'Сетка страниц',exact:true}).click();assert.equal(await page.locator('.page-overview button').count(),3);await page.locator('.page-overview button').nth(2).click();assert.equal(await page.locator('#studio-page-workflow-page-2').evaluate(el=>el===document.activeElement),true);
 ready=true;await page.locator('.processing-summary').getByRole('button',{name:'Обновить состояние'}).click();await page.waitForFunction(()=>document.querySelector('.processing-summary')?.textContent.includes('Готово: 3'));
 await page.locator('.draft-steps a[href="#draft-preview"]').click();assert.equal(await page.getByRole('button',{name:'Отправить на проверку',exact:true}).isEnabled(),false,'saved metadata must be reloaded explicitly when it changed externally');
 await page.reload();await page.locator('.preflight').waitFor();assert.equal(await page.getByRole('button',{name:'Отправить на проверку',exact:true}).isEnabled(),true);
 const href=await page.getByRole('link',{name:'Предпросмотр',exact:true}).getAttribute('href');const url=new URL(href,fixture.app_url);assert.equal(url.searchParams.get('preview'),'workflow-revision');assert.match(url.searchParams.get('returnTo'),new RegExp('^/studio/projects/'+fixture.project_id+'\\?revision=workflow-revision&step=draft-preview&position=\\d+$'));
 for(const locale of ['kk','ru'])for(const theme of ['dark','light'])for(const width of [320,360,1280]){await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await page.setViewportSize({width,height:850});await page.reload();await page.locator('.preflight').waitFor();await page.evaluate(value=>document.documentElement.dataset.theme=value,theme);await page.locator('.draft-steps a[href="#draft-pages"]').click();await page.locator('#draft-pages button[aria-pressed]').click();assert.equal(await page.locator('.page-overview button').count(),3);assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));}
 }finally{await context.close();await browser.close();}
});
test('project return restores applied criteria and loaded pages without applying a draft query',{timeout:45000},async()=>{
 const {browser,context}=await session();try{const page=await context.newPage(),calls=[];
 await page.route('**/api/v1/studio/projects?*',route=>{const url=new URL(route.request().url());calls.push(url);const second=!!url.searchParams.get('cursor');return route.fulfill({json:{items:Array.from({length:20},(_,index)=>({id:index?'context-project-'+(second?'next-':'first-')+index:fixture.project_id,title_name:index?'Context project '+index:(second?'Second project':'First project'),team_name:'Synthetic team',team_id:'synthetic-team',next_revision_id:'work-id',next_revision_state:'processing'})),next_cursor:second?'':'100',total:2}});});
 await page.goto('/studio/projects?q=Applied&pages=2&position=100');await page.getByText('Second project',{exact:true}).first().waitFor();assert.equal(calls.length,2);assert.equal(calls[1].searchParams.get('q'),'Applied');
 await page.locator('.project-search input').fill('Unapplied draft');await page.evaluate(()=>window.scrollTo(0,600));await page.waitForFunction(()=>window.scrollY>=500);await page.waitForTimeout(50);const href=await page.locator('a.manage').first().getAttribute('href');const target=new URL(href,fixture.app_url);assert.equal(target.searchParams.get('revision'),'work-id');const back=new URL(target.searchParams.get('projectsReturn'),fixture.app_url);assert.equal(back.searchParams.get('q'),'Applied');assert.equal(back.searchParams.get('pages'),'2');assert.ok(Number(back.searchParams.get('position'))>=500);
 await page.goto(back.href);await page.getByText('Second project',{exact:true}).first().waitFor();assert.equal(await page.locator('.project-search input').inputValue(),'Applied');await page.waitForFunction(()=>window.scrollY>=500);
 }finally{await context.close();await browser.close();}
});
test('reader preview returns to the exact authorized workspace version and step',{timeout:45000},async()=>{
 const {browser,context}=await session();try{
 const response=await context.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions?offset=0');assert.equal(response.status(),200);
 const revision=(await response.json()).items.find(item=>item.release_id===fixture.release_id&&item.state==='published');assert.ok(revision,'published fixture revision required');
 const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id+'?revision='+revision.id+'&step=draft-preview');await page.locator('#draft-preview').waitFor({state:'visible'});
 const preview=page.getByRole('link',{name:'Предпросмотр',exact:true});const expected=new URL(await preview.getAttribute('href'),fixture.app_url).searchParams.get('returnTo');await preview.click();await page.locator('.preview-note a').waitFor();assert.equal(await page.locator('.preview-note a').getAttribute('href'),expected);await page.getByRole('link',{name:'Вернуться к работе',exact:true}).click();await page.locator('#draft-preview').waitFor({state:'visible'});assert.equal(new URL(page.url()).searchParams.get('revision'),revision.id);assert.equal(await page.locator('.revision-option[aria-current="true"]').count(),1);
 await page.reload();await page.locator('#draft-preview').waitFor({state:'visible'});assert.equal(new URL(page.url()).searchParams.get('step'),'draft-preview');
 await page.route('**/api/v1/studio/revisions/unavailable-workflow',route=>route.fulfill({status:404,json:{error:{code:'not_found'}}}));await page.goto('/studio/projects/'+fixture.project_id+'?revision=unavailable-workflow&step=draft-preview');await page.getByRole('alert').first().waitFor();assert.match(await page.getByRole('alert').first().innerText(),/Запрошенная версия недоступна/);assert.equal(await page.locator('.draft-panel h2').count(),1);
 }finally{await context.close();await browser.close();}
});
