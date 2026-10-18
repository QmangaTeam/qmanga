import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {mkdir,readFile} from 'node:fs/promises';
import {join} from 'node:path';
const fixture=process.env.QMANGA_FIXTURE_FILE?JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')):null;
const baseURL=fixture?.app_url||process.env.QMANGA_TEST_URL||'http://localhost:8080';
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local typography checks only');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});

test('one local variable font serves Kazakh/Russian and keeps reading layouts within the viewport',{timeout:60000},async()=>{
 for(const locale of ['kk','ru'])for(const theme of ['light','dark']){
  const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:baseURL}]);
   await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   const page=await context.newPage();const fonts=[];
   page.on('response',response=>{if(response.request().resourceType()==='font')fonts.push(response);});
   await page.goto('/');
   const result=await page.evaluate(async()=>{
    const regular=await document.fonts.load('400 16px Onest','ӘҒҚҢӨҰҮҺІ әғқңөұүһі Русский 0123456789');
    const bold=await document.fonts.load('700 16px Onest','Қазақша Русский');
    return {regular:regular.length,bold:bold.length,family:getComputedStyle(document.body).fontFamily};
   });
   assert.equal(result.regular,1,'Kazakh and Russian use the bundled font');assert.equal(result.bold,1);
   assert.ok(result.family.startsWith('Onest'));
   assert.equal(fonts.length,1,'one variable font request covers both weights and scripts');
   assert.equal(new URL(fonts[0].url()).origin,new URL(baseURL).origin);
   assert.ok((await fonts[0].body()).length<90*1024,'font stays below the 90 KiB transfer budget');
   const routes=['/','/catalog','/menu','/register',...(fixture?['/read/'+fixture.release_id]:[])];
   for(const route of routes){
    if(new URL(page.url()).pathname!==route)await page.goto(route);
    for(const width of [320,360,1280]){
     await page.setViewportSize({width,height:800});
     assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,`${route} ${locale} ${theme} ${width}`);
    }
    if(process.env.QMANGA_SCREENSHOT_DIR){
     await page.setViewportSize({width:360,height:800});await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});
     await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`font-${route.split('/')[1]||'home'}-${locale}-${theme}.png`)});
    }
   }
  }finally{await context.close();}
 }
});

test('interface remains usable when the font download fails',{timeout:15000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:320,height:800}});
 try{
  await context.route('**/*.woff2',route=>route.abort());
  const page=await context.newPage();await page.goto('/register');
  await page.getByRole('heading',{level:1}).waitFor();
  assert.equal(await page.getByRole('textbox').first().isVisible(),true);
  assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
 }finally{await context.close();}
});
