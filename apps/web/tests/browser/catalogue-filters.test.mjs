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
for(const locale of ['kk','ru'])for(const theme of ['light','dark'])test(`catalogue filters follow the reference panel and apply instantly: ${locale}/${theme}`,{timeout:60000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:1280,height:800},reducedMotion:'reduce'});
 let page;const browserErrors=[];
 const triggerName=locale==='ru'?'Фильтры':'Сүзгілер';
 const closeName=locale==='ru'?'Закрыть фильтры':'Сүзгілерді жабу';
 const resetName=locale==='ru'?'Сбросить фильтры':'Сүзгілерді тазарту';
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
  assert.equal(await aside.locator('input[type="checkbox"][value="18"]').isChecked(),true);
  assert.equal(await aside.locator('input[type="checkbox"][value="manhwa"]').isChecked(),true);
  assert.equal(await aside.getByRole('button',{name:locale==='ru'?'Применить':'Қолдану',exact:true}).count(),0,'the reference panel has no apply button');
  const desktopScroll=aside.locator('.panel-scroll');
  assert.equal(await desktopScroll.evaluate(el=>getComputedStyle(el).overflowY),'auto');
  assert.equal(await desktopScroll.evaluate(el=>el.scrollHeight>el.clientHeight+1),true,'desktop filter panel must scroll internally instead of clipping its content');
  const reach=await aside.evaluate(el=>{
   const ps=el.querySelector('.panel-scroll');
   ps.scrollTop=ps.scrollHeight;
   const clip=el.querySelector('.panel-clip').getBoundingClientRect();
   const heads=[...el.querySelectorAll('.filter-section h3')];
   const last=heads[heads.length-1].getBoundingClientRect();
   return {moved:ps.scrollTop>0,inside:last.y>=clip.y-1&&last.y+last.height<=clip.y+clip.height+1};
  });
  assert.equal(reach.moved,true,'desktop filter panel scrolls to the end');
  assert.equal(reach.inside,true,'the last filter section is reachable inside the panel');
  for(const width of [320,360]){
   await page.setViewportSize({width,height:800});
   await page.goto(url);
   const trigger=page.getByRole('button',{name:triggerName,exact:true});const beforeOpen=await page.locator('.catalogue-results').boundingBox();await trigger.click();
   const dialog=page.getByRole('dialog',{name:triggerName,exact:true});await dialog.waitFor();
   assert.equal(await aside.isVisible(),false);
   const afterOpen=await page.locator('.catalogue-results').boundingBox();assert.equal(afterOpen.width,beforeOpen.width,'opening filters must not resize the result grid');
   assert.equal(afterOpen.x,beforeOpen.x);
   const panel=await dialog.boundingBox();const viewport=await page.evaluate(()=>({width:innerWidth,height:innerHeight}));assert.ok(panel.x<=0.5&&panel.y<=0.5&&panel.width>=viewport.width-1&&panel.height>=viewport.height-1,'mobile filters fill the screen');
   assert.equal(await dialog.locator('input[type="checkbox"][value="18"]').isChecked(),true);
   assert.equal(await dialog.locator('input[type="checkbox"][value="manhwa"]').isChecked(),true);
   assert.equal(await dialog.evaluate(el=>el.contains(document.activeElement)),true);
   assert.equal(await dialog.evaluate(el=>getComputedStyle(el).animationDuration),'0s','reduced motion disables the slide');
   const close=dialog.getByRole('button',{name:closeName,exact:true});
   const box=await close.boundingBox();assert.ok(box.y>=0&&box.y+box.height<=800,'the bottom close control stays above the screen edge');
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
   await dialog.locator('input[type="checkbox"][value="12"]').check();
   await page.waitForURL(value=>value.searchParams.getAll('age_rating').includes('12'));
   assert.equal(await dialog.locator('input[type="checkbox"][value="12"]').isChecked(),true,'instant apply keeps the dialog state in sync');
   await dialog.locator('input[type="checkbox"][value="18"]').uncheck();
   await page.waitForURL(value=>!value.searchParams.getAll('age_rating').includes('18'));
   assert.deepEqual(new URL(page.url()).searchParams.getAll('age_rating'),['12']);
   await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
   assert.equal(await trigger.evaluate(el=>el===document.activeElement),true,'focus returns to the filter trigger');
   await trigger.click();assert.equal(await dialog.locator('input[type="checkbox"][value="12"]').isChecked(),true,'reopening restores the committed URL state');
   await dialog.getByRole('button',{name:closeName,exact:true}).click();await dialog.waitFor({state:'hidden'});
   if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`filters-${locale}-${theme}-${width}.png`)});}
  }
  await page.setViewportSize({width:1280,height:800});
  assert.equal(await aside.isVisible(),true);
  assert.equal(await aside.locator('input[type="checkbox"][value="12"]').isChecked(),true);
  await aside.locator('input[type="checkbox"][value="manhwa"]').uncheck();
  await page.waitForURL(value=>!value.searchParams.has('kind'));
  await aside.locator('input[type="checkbox"][value="manga"]').check();
  await page.waitForURL(value=>value.searchParams.get('kind')==='manga');
  assert.equal(await aside.locator('input[type="checkbox"][value="manhwa"]').isChecked(),false,'a single kind is selected');
  const chapters=aside.locator('input[name="chapters_min"]');
  await chapters.fill('20');await chapters.press('Enter');
  await page.waitForURL(value=>value.searchParams.get('chapters_min')==='20');
  await aside.getByRole('button',{name:'21–100',exact:true}).click();
  await page.waitForURL(value=>value.searchParams.get('chapters_min')==='21'&&value.searchParams.get('chapters_max')==='100');
  const year=aside.locator('input[name="year_min"]');
  await year.fill('2020');await year.press('Enter');
  await page.waitForURL(value=>value.searchParams.get('year_min')==='2020');
  const applied=new URL(page.url());
  assert.equal(applied.searchParams.get('q'),'DrawerFixture');
  assert.equal(applied.searchParams.get('sort'),'name');
  assert.equal(applied.searchParams.has('cursor'),false);
  assert.equal(await aside.locator('input[name="chapters_max"]').inputValue(),'100');
  await aside.getByRole('button',{name:resetName,exact:true}).click();
  await page.waitForURL(value=>value.pathname==='/catalog'&&!value.searchParams.has('age_rating')&&!value.searchParams.has('kind')&&!value.searchParams.has('chapters_min'));
  assert.equal(new URL(page.url()).searchParams.get('q'),'DrawerFixture','reset keeps the global search');
  assert.equal(new URL(page.url()).searchParams.get('sort'),'name','reset keeps the result sorting');
  assert.equal(await aside.locator('input[type="checkbox"]:checked').count(),0);
  assert.equal(await aside.locator('input[name="year_min"]').inputValue(),'');
  assert.equal(await page.evaluate(()=>getComputedStyle(document.body).overflow),'visible');
  assert.deepEqual(browserErrors.filter(entry=>entry.kind==='pageerror'),[],'the filter panel has no page errors');
  if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`filters-${locale}-${theme}-desktop.png`)});}
 }catch(error){
  const directory=process.env.QMANGA_SCREENSHOT_DIR||join(process.cwd(),'tmp','browser-failures');
  await mkdir(directory,{recursive:true});
  const file=join(directory,`filters-failure-${locale}-${theme}`);
  let state;
  try{state=await page?.evaluate(()=>({url:location.href,ready:document.readyState,asideCount:document.querySelectorAll('aside.catalogue-filters').length,dialogs:document.querySelectorAll('dialog[open]').length,main:document.querySelector('main')?.innerText}));}catch(cause){state={diagnosticError:String(cause)};}
  await writeFile(file+'.json',JSON.stringify({error:String(error),browserErrors,state},null,2));
  try{await page?.screenshot({path:file+'.png',timeout:3000});}catch{/* Keep the original failure if a crashed page cannot be captured. */}
  throw error;
 }finally{await context.close();}
});
