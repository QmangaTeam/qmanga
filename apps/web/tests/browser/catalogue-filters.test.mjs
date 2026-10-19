import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {readFile,mkdir,writeFile} from 'node:fs/promises';
import {join} from 'node:path';
const fixture=process.env.QMANGA_FIXTURE_FILE?JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')):null;
const baseURL=fixture?.app_url||process.env.QMANGA_TEST_URL||'http://localhost:8080';
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local filter regression only');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
for(const locale of ['kk','ru'])for(const theme of ['light','dark'])test(`catalogue sidebar and drawer preserve committed URL state: ${locale}/${theme}`,{timeout:30000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:1280,height:800},reducedMotion:'reduce'});
 let page;const browserErrors=[];
 try{
  await context.addCookies([{name:'qmanga_locale',value:locale,url:baseURL}]);await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
  page=await context.newPage();page.setDefaultTimeout(8000);
  page.on('pageerror',error=>browserErrors.push({kind:'pageerror',message:error.message,stack:error.stack}));
  page.on('console',entry=>{if(entry.type()==='error')browserErrors.push({kind:'console',message:entry.text()});});
  const url='/catalog?q=DrawerFixture&age_rating=18&kind=manhwa&sort=name';
  await page.goto(url);
  assert.equal(await page.locator('.catalogue-search,#catalogue-query').count(),0,'only global search remains');
  const aside=page.locator('aside.catalogue-filters');
  assert.equal(await aside.isVisible(),true,'desktop filters must be permanently visible on the right');
  const sidebar=await aside.boundingBox(),results=await page.locator('.catalogue-results').boundingBox();assert.ok(sidebar.x>results.x+results.width-1);
  assert.equal(await aside.locator('select[name="age_rating"]').inputValue(),'18');
  for(const width of [320,360]){
   await page.setViewportSize({width,height:800});
   const trigger=page.getByRole('button',{name:locale==='ru'?'Фильтры':'Сүзгілер',exact:true});const beforeOpen=await page.locator('.catalogue-results').boundingBox();await trigger.click();
   const dialog=page.getByRole('dialog',{name:locale==='ru'?'Фильтры каталога':'Каталог сүзгілері',exact:true});await dialog.waitFor();
   assert.equal(await aside.isVisible(),false);
   const afterOpen=await page.locator('.catalogue-results').boundingBox();assert.equal(afterOpen.width,beforeOpen.width,'opening filters must not resize the result grid');
   assert.equal(afterOpen.x,beforeOpen.x);
   assert.equal(await dialog.locator('select[name="age_rating"]').inputValue(),'18');
   assert.equal(await dialog.locator('select[name="kind"]').inputValue(),'manhwa');
   assert.equal(await dialog.evaluate(el=>el.contains(document.activeElement)),true);
   assert.equal(await dialog.evaluate(el=>getComputedStyle(el).animationDuration),'0s','reduced motion disables the slide');
   const apply=dialog.getByRole('button',{name:locale==='ru'?'Применить':'Қолдану',exact:true});
   const box=await apply.boundingBox();assert.ok(box.y>=0&&box.y+box.height<=800,'Apply remains above the screen edge and bottom navigation');
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
   await dialog.locator('select[name="age_rating"]').selectOption('12');
   await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
   assert.equal(await trigger.evaluate(el=>el===document.activeElement),true);
   await trigger.click();assert.equal(await dialog.locator('select[name="age_rating"]').inputValue(),'18','reopening restores URL selection, not canceled edits');
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`filters-${locale}-${theme}-${width}.png`)});}
   await dialog.getByRole('button',{name:locale==='ru'?'Закрыть фильтры':'Сүзгілерді жабу',exact:true}).click();await dialog.waitFor({state:'hidden'});
  }
  await page.goto(url+'&cursor=stale');
  const trigger=page.getByRole('button',{name:locale==='ru'?'Фильтры':'Сүзгілер',exact:true});await trigger.click();
  const dialog=page.getByRole('dialog');
  await dialog.locator('select[name="age_rating"]').selectOption('12');await dialog.locator('select[name="kind"]').selectOption('manga');
  await dialog.getByRole('button',{name:locale==='ru'?'Применить':'Қолдану',exact:true}).click();
  await page.waitForURL(value=>value.pathname==='/catalog'&&value.searchParams.get('age_rating')==='12');
  assert.equal(new URL(page.url()).searchParams.get('q'),'DrawerFixture');assert.equal(new URL(page.url()).searchParams.get('kind'),'manga');assert.equal(new URL(page.url()).searchParams.has('cursor'),false);
  await trigger.click();assert.equal(await dialog.locator('select[name="age_rating"]').inputValue(),'12');
  await page.setViewportSize({width:1280,height:800});await dialog.waitFor({state:'hidden'});assert.equal(await aside.isVisible(),true);
  assert.notEqual(await page.evaluate(()=>getComputedStyle(document.body).overflow),'hidden');
  if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`filters-${locale}-${theme}-desktop.png`)});
  await page.setViewportSize({width:668,height:800});await trigger.click();await page.mouse.click(8,200);await dialog.waitFor({state:'hidden'});
 }catch(error){
  const directory=process.env.QMANGA_SCREENSHOT_DIR||join(process.cwd(),'tmp','browser-failures');
  await mkdir(directory,{recursive:true});
  const file=join(directory,`filters-failure-${locale}-${theme}`);
  let state;
  try{state=await page?.evaluate(()=>({url:location.href,ready:document.readyState,asideCount:document.querySelectorAll('aside.catalogue-filters').length,main:document.querySelector('main')?.innerText}));}catch(cause){state={diagnosticError:String(cause)};}
  await writeFile(file+'.json',JSON.stringify({error:String(error),browserErrors,state},null,2));
  try{await page?.screenshot({path:file+'.png',timeout:3000});}catch{/* Keep the original failure if a crashed page cannot be captured. */}
  throw error;
 }finally{await context.close();}
});
