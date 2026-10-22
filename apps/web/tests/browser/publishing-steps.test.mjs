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

test('chapter steps retain edits, recover failed saves and expose only the chosen section',{timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});
  assert.equal(login.status(),200);
  const page=await context.newPage();
  await page.goto('/studio/projects/'+fixture.project_id);
  await page.getByRole('button',{name:'Добавить главу',exact:true}).click();
  const label='Steps-'+Date.now();
  await page.getByRole('textbox',{name:'Номер или обозначение главы'}).fill(label);
  await page.getByRole('button',{name:'Создать черновик',exact:true}).click();
  await page.locator('.draft-panel h2').filter({hasText:label}).waitFor();
  assert.equal(await page.locator('#draft-details').isVisible(),false,'source fields should wait for the details step');
  assert.equal(await page.locator('#draft-preview').isVisible(),false,'submission should wait for the review step');
  assert.equal(await page.locator('.chapter-sidebar form').isVisible(),false,'chapter choice collapses after opening');
  await page.getByRole('button',{name:'Выбрать главу',exact:true}).click();
  await page.locator('.chapter-sidebar select').first().selectOption({label:'Добавить новую главу'});
  assert.equal(await page.getByRole('textbox',{name:'Номер или обозначение главы'}).inputValue(),'','saved chapter input is cleared after creation');
  await page.locator('.chapter-sidebar select').first().selectOption({label:'Выбрать существующую'});
  await page.locator('.chapter-switch-dialog').getByRole('button',{name:'Закрыть',exact:true}).click();
  await page.locator('#draft-pages input[type="file"]').first().setInputFiles(join(dirname(process.env.QMANGA_FIXTURE_FILE),'pages','page-01.png'));
  const next=page.getByRole('button',{name:'Далее: сведения',exact:true});
  await next.waitFor();
  await page.waitForFunction(()=>[...document.querySelectorAll('button')].some(b=>b.textContent==='Далее: сведения'&&!b.disabled));
  await next.click();
  assert.equal(await page.locator('#draft-pages').isVisible(),false);
  const source=page.getByRole('textbox',{name:'Источник и основание права публикации'});
  await source.fill('Synthetic local pages; original geometric test image.');
  await page.getByRole('textbox',{name:'Кредиты: авторы и переводчики'}).fill('Qmanga local tooling');
  await page.getByRole('checkbox',{name:'Подтверждаю наличие права на публикацию этого материала',exact:true}).check();
  await page.locator('.draft-steps a[href="#draft-pages"]').click();
  await page.locator('.draft-steps a[href="#draft-details"]').click();
  assert.equal(await source.inputValue(),'Synthetic local pages; original geometric test image.');
  await page.route('**/api/v1/studio/revisions/*/attribution',route=>route.fulfill({status:500,contentType:'application/json',body:JSON.stringify({error:{code:'internal'}})}));
  await page.getByRole('button',{name:'Сохранить и проверить',exact:true}).click();
  await page.getByRole('alert').first().waitFor();
  assert.equal(await source.isVisible(),true,'failed save must leave current step open');
  assert.equal(await source.inputValue(),'Synthetic local pages; original geometric test image.');
  await page.unroute('**/api/v1/studio/revisions/*/attribution');
  await page.getByRole('button',{name:'Сохранить и проверить',exact:true}).click();
  await page.locator('#draft-preview').waitFor({state:'visible'});
  assert.equal(await page.locator('#draft-details').isVisible(),false);
  assert.equal(await page.getByRole('button',{name:'Отправить на проверку',exact:true}).isEnabled(),true);
  for(const width of [320,360,1280]){
   await page.setViewportSize({width,height:800});
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'steps fit at '+width);
   for(const id of ['draft-pages','draft-details','draft-preview'])assert.equal(await page.locator('#'+id).isVisible(),id==='draft-preview','responsive visibility '+id);
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'publishing-steps-'+width+'.png'),fullPage:true});}
  }
  await page.getByRole('button',{name:'Удалить черновик',exact:true}).first().click();
  await page.locator('.discard-dialog').getByRole('button',{name:'Удалить черновик',exact:true}).click();
  await page.getByText('Черновик удалён',{exact:true}).waitFor();
  const published=await context.request.get('/api/v1/releases/'+fixture.release_id+'/manifest');
  assert.equal(published.status(),200);
 }finally{await context.close();}
});

test('step navigation fits Kazakh and Russian in both themes with keyboard access',{timeout:60000},async()=>{
 for(const locale of ['kk','ru'])for(const theme of ['light','dark']){
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:320,height:800},reducedMotion:'reduce'});
  try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);
   await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});
   assert.equal(login.status(),200);
   const page=await context.newPage();
   await page.goto('/studio/projects/'+fixture.project_id);
  await page.locator('.version-row a').first().click();
  await page.locator('.draft-panel h2').waitFor();
   const rail=page.locator('.draft-steps');await rail.waitFor();
   const details=rail.locator('a[href="#draft-details"]');
   await details.press('Enter');
   await page.locator('#draft-details').waitFor({state:'visible'});
   assert.equal(await page.locator('#draft-pages').isVisible(),false);
   assert.equal(await page.locator('#draft-preview').isVisible(),false);
   assert.equal(await page.locator('#draft-details h3').evaluate(el=>el===document.activeElement),true,'focus follows selected step');
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:800});
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,locale+' '+theme+' '+width);
    const links=await rail.locator('a').evaluateAll(elements=>elements.map(el=>el.getBoundingClientRect().height));
    assert.ok(links.every(height=>height>=44),'step targets are at least 44px');
    if(width>=1000){assert.equal(await page.locator('#draft-details').isVisible(),true);await rail.locator('a[href="#draft-preview"]').press('Enter');assert.equal(await page.locator('#draft-preview h3').evaluate(el=>el===document.activeElement),true);await details.click();}
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`steps-${locale}-${theme}-${width}.png`),fullPage:false});}
   }
  }finally{await context.close();}
 }
});

test('new chapter input asks before closing and survives failure and cancelled dismissal',{timeout:30000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});try{await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id);await page.getByRole('button',{name:'Добавить главу',exact:true}).click();const modal=page.locator('.new-chapter-dialog');await modal.getByRole('textbox',{name:'Номер или обозначение главы',exact:true}).fill('Unsaved number');await modal.getByRole('textbox',{name:'Название · необязательно',exact:true}).fill('Unsaved title');await page.route('**/api/v1/studio/projects/*/chapters',route=>route.fulfill({status:500,json:{error:{code:'internal'}}}));await modal.getByRole('button',{name:'Создать черновик',exact:true}).click();await modal.getByRole('alert').waitFor();page.once('dialog',d=>d.dismiss());await page.keyboard.press('Escape');assert.equal(await modal.evaluate(el=>el.open),true);assert.equal(await modal.getByRole('textbox',{name:'Номер или обозначение главы',exact:true}).inputValue(),'Unsaved number');page.once('dialog',d=>d.accept());await page.keyboard.press('Escape');assert.equal(await modal.evaluate(el=>el.open),false);}finally{await context.close();}
});
