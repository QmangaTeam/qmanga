import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {dirname,join} from 'node:path';
import {chromium} from 'playwright';

if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Disposable local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});

test('artwork changes require moderation, keep evidence private and support a cover fallback banner', {timeout:150000},async()=>{
 const owner=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});
 const staff=await browser.newContext({baseURL:fixture.app_url,viewport:{width:1280,height:900}});
 const guest=await browser.newContext({baseURL:fixture.app_url});
 let pending='';let actor='',moderator='';
 async function protectedGet(context,path,id){return context.request.get(path,{headers:{'X-Qmanga-Account':id}});}
 async function review(id,decision,reason=''){const csrf=await (await staff.request.get('/api/v1/auth/csrf')).json();return staff.request.post('/api/v1/moderation/title-changes/'+id+'/review',{headers:{Origin:fixture.app_url,'X-CSRF-Token':csrf.csrf_token,'X-Qmanga-Account':moderator},data:{decision,reason}});}
 try{
  for(const context of [owner,staff])await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  let login=await owner.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);actor=(await login.json()).user.id;
  login=await staff.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}});assert.equal(login.status(),200);moderator=(await login.json()).user.id;
  const before=await (await owner.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();
  const revisions=await (await owner.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions')).json();const count=revisions.items.reduce((n,x)=>n+x.pages.length,0);
  const csrf=await (await owner.request.get('/api/v1/auth/csrf')).json();
  for(const endpoint of ['cover','cover-file']){const response=await owner.request.put('/api/v1/studio/projects/'+fixture.project_id+'/'+endpoint,{headers:{Origin:fixture.app_url,'X-CSRF-Token':csrf.csrf_token},data:{page_id:''}});assert.equal(response.status(),409);assert.equal((await response.json()).error.code,'moderation_required');}
  const page=await owner.newPage();await page.goto('/studio/projects/'+fixture.project_id);await page.locator('.project-navigation').getByRole('link',{name:'Сведения о тайтле',exact:true}).click();
  await page.getByRole('heading',{name:'Изменить данные проекта',exact:true}).waitFor();await page.getByRole('textbox',{name:'Название',exact:true}).waitFor();
  assert.equal(await page.locator('input[name="slug"]').count(),0);
  const file=join(dirname(process.env.QMANGA_FIXTURE_FILE),'pages','page-01.png');
  const cover=page.locator('.artwork').filter({has:page.getByRole('heading',{name:'Обложка',exact:true})});const banner=page.locator('.artwork').filter({has:page.getByRole('heading',{name:'Баннер',exact:true})});
  await cover.getByRole('combobox').selectOption('replace');await cover.locator('input[type=file]').setInputFiles(file);
  await banner.getByRole('combobox').selectOption('replace');await banner.locator('input[type=file]').setInputFiles(file);
  await page.getByRole('textbox',{name:'Что и почему нужно изменить?'}).fill('Обложка и баннер по официальному источнику');
  await page.getByLabel('Изображения доказательств', {exact:false}).setInputFiles(file);
  for(const width of [320,360,1280]){await page.setViewportSize({width,height:900});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'form overflow '+width);if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'title-change-form-'+width+'.png'),fullPage:true});}}
  await page.getByRole('button',{name:'Отправить на модерацию',exact:true}).click();await page.getByText('Заявка отправлена на модерацию.',{exact:true}).waitFor();
  const history=await (await protectedGet(owner,'/api/v1/studio/projects/'+fixture.project_id+'/title-changes',actor)).json();const request=history.items.find(x=>x.status==='pending');assert.ok(request);pending=request.id;
  const unchanged=await (await owner.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();assert.equal(unchanged.cover_id,before.cover_id);
  let detail=await (await protectedGet(owner,'/api/v1/title-changes/'+pending,actor)).json();const evidence=detail.request.assets.find(a=>a.purpose==='evidence'),proposedCover=detail.request.assets.find(a=>a.purpose==='cover'),proposedBanner=detail.request.assets.find(a=>a.purpose==='banner');assert.ok(evidence&&proposedCover&&proposedBanner);
  assert.equal((await guest.request.get(evidence.url)).status(),404);assert.equal((await guest.request.get(proposedCover.url)).status(),404);const privateImage=await owner.request.get(evidence.url);assert.equal(privateImage.status(),200);assert.match(privateImage.headers()['cache-control'],/no-store/);
  const staffPage=await staff.newPage();await staffPage.goto('/moderation/title-changes/'+pending);await staffPage.getByRole('heading',{name:'Решение модератора'}).waitFor();await staffPage.getByRole('button',{name:'Одобрить',exact:true}).click();await staffPage.getByRole('heading',{name:'Решение',exact:true}).waitFor();pending='';
  assert.equal((await guest.request.get(proposedCover.url)).status(),200);assert.equal((await guest.request.get(proposedBanner.url)).status(),200);assert.equal((await guest.request.get(evidence.url)).status(),404);
  const changed=await (await owner.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();assert.equal(changed.cover_id,proposedCover.id);
  const publicContext=await (await protectedGet(owner,'/api/v1/studio/projects/'+fixture.project_id+'/title-change-context',actor)).json();
  await page.goto('/titles/'+publicContext.slug);await page.locator('.title-banner img').waitFor();assert.equal(await page.locator('.title-banner img').getAttribute('src'),proposedBanner.url);
  for(const theme of ['light','dark']){await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);for(const width of [320,360,1280]){await page.setViewportSize({width,height:900});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'banner overflow '+width);if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'title-banner-'+theme+'-'+width+'.png'),fullPage:true});}}
  await page.route('**'+proposedBanner.url,route=>route.abort());await page.reload();await page.waitForFunction(url=>document.querySelector('.title-banner img')?.getAttribute('src')===url,proposedCover.url);await page.unroute('**'+proposedBanner.url);
  const after=await (await owner.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions')).json();assert.equal(after.items.reduce((n,x)=>n+x.pages.length,0),count);
  // Removing a separate banner is also a reviewed proposal; approval restores the cover fallback.
  const proposal={base_version:publicContext.base_version,data:publicContext.snapshot.data,reason:'Использовать обложку вместо отдельного баннера',links:['https://example.org/official'],cover_action:'keep',banner_action:'remove',cover_page_id:''};
  const submitted=await owner.request.post('/api/v1/studio/projects/'+fixture.project_id+'/title-changes',{headers:{Origin:fixture.app_url,'X-CSRF-Token':csrf.csrf_token,'X-Qmanga-Account':actor},multipart:{proposal:JSON.stringify(proposal)}});assert.equal(submitted.status(),201);pending=(await submitted.json()).id;
  assert.equal((await review(pending,'approve')).status(),204);pending='';await page.reload();await page.locator('.title-banner img').waitFor();assert.equal(await page.locator('.title-banner img').getAttribute('src'),proposedCover.url);
  await owner.addCookies([{name:'qmanga_locale',value:'kk',url:fixture.app_url}]);await page.goto('/studio/projects/'+fixture.project_id+'/metadata');await page.getByRole('textbox',{name:'Атауы',exact:true}).waitFor();await page.setViewportSize({width:320,height:850});await page.evaluate(()=>document.documentElement.dataset.theme='dark');assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'Kazakh form fits phone');if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'title-change-form-kk-dark-320.png'),fullPage:true});
 }finally{if(pending&&moderator)await review(pending,'reject','Cleanup incomplete browser scenario');await owner.close();await staff.close();await guest.close();}
});

test('accidental draft can be cancelled then discarded without touching the published chapter', {timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:1280,height:800}});
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});
  assert.equal(login.status(),200);
  const label='Accidental-'+Date.now();
  const page=await context.newPage();
  try{
   await page.goto('/studio/projects/'+fixture.project_id);
   await page.getByRole('button',{name:'Добавить главу',exact:true}).click();
   await page.getByRole('textbox',{name:'Номер или обозначение главы'}).fill(label);
   await page.getByRole('button',{name:'Создать черновик',exact:true}).click();
   await page.locator('.draft-panel h2').filter({hasText:label}).waitFor();
   const rail=page.locator('.draft-steps');
   assert.equal(await rail.locator('a').count(),3);
   const railX=await rail.evaluate(element=>element.getBoundingClientRect().left);
   const panelX=await page.locator('.draft-panel').evaluate(element=>element.getBoundingClientRect().left);
   assert.ok(Math.abs(railX-panelX)<2,'editor tabs align with the working surface');
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'draft-rail-1280.png'),fullPage:false});}
   await rail.locator('a[href="#draft-details"]').click();
   await page.waitForFunction(()=>document.querySelector('.draft-steps a[href="#draft-details"]')?.getAttribute('aria-current')==='location');
   await page.getByRole('button',{name:'Удалить черновик'}).first().click();
   const dialog=page.locator('.discard-dialog');
   await dialog.getByRole('heading',{name:'Удалить черновик?'}).waitFor();
   await dialog.getByRole('button',{name:'Отмена'}).click();
   assert.equal(await page.locator('.draft-panel h2').filter({hasText:label}).count(),1);
   if(process.env.QMANGA_SCREENSHOT_DIR){await page.setViewportSize({width:320,height:750});await page.evaluate(()=>scrollTo(0,0));await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'draft-steps-320.png'),fullPage:false});}
   await page.getByRole('button',{name:'Удалить черновик'}).first().click();
   await dialog.getByRole('button',{name:'Удалить черновик'}).click();
   await page.getByText('Черновик удалён').waitFor();
   const revisions=await (await context.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions')).json();
   assert.equal(revisions.items.some(item=>item.chapter_label===label),false);
   const published=await context.request.get('/api/v1/releases/'+fixture.release_id+'/manifest');
   assert.equal(published.status(),200);
   await page.setViewportSize({width:320,height:750});
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
  }finally{await page.close();}
 }finally{await context.close();}
});
