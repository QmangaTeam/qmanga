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

test('editor uploads a separate title cover without changing chapter pages', {timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});
  assert.equal(login.status(),200);
  const before=await (await context.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();
  const revisions=await (await context.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions')).json();
  const pageCount=revisions.items.reduce((sum,revision)=>sum+revision.pages.length,0);
  const page=await context.newPage();
  try{
   await page.goto('/studio/projects/'+fixture.project_id);
   const panel=page.locator('.cover-panel');await panel.getByRole('heading',{name:'Обложка тайтла'}).waitFor();
   assert.equal(await panel.locator('input[type="file"]').count(),1);
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,'cover-studio-360.png'),fullPage:true});}
   const imagePath=join(dirname(process.env.QMANGA_FIXTURE_FILE),'pages','page-01.png');
   await panel.locator('input[type="file"]').setInputFiles(imagePath);
   await page.waitForFunction(old=>{const image=document.querySelector('.cover-panel img');return !!image&&image.getAttribute('src')!=='/api/v1/media/'+old;},before.cover_id);
   const updated=await (await context.request.get('/api/v1/studio/projects/'+fixture.project_id)).json();
   assert.notEqual(updated.cover_id,before.cover_id);
   const after=await (await context.request.get('/api/v1/studio/projects/'+fixture.project_id+'/revisions')).json();
   assert.equal(after.items.reduce((sum,revision)=>sum+revision.pages.length,0),pageCount);
   await page.setViewportSize({width:320,height:750});
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'cover controls must fit at 320px');
  }finally{await page.close();}
 }finally{await context.close();}
});
