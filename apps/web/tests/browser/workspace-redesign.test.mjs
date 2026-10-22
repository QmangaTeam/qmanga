import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('shared admin shell retains nine destinations, account access and responsive palette',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
  const context=await browser.newContext({baseURL:fixture.app_url,reducedMotion:'reduce'});try{
   await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}});assert.equal(login.status(),200);
   const page=await context.newPage();await page.goto('/admin');await page.locator('.workspace-topbar').waitFor();
   for(const width of [320,360,1280]){await page.setViewportSize({width,height:900});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'admin fits viewport');
    const nav=page.locator(width<1000?'.workspace-drawer nav':'.workspace-sidebar nav');
    if(width<1000){await page.waitForFunction(()=>!document.querySelector('.workspace-menu')?.disabled);await page.locator('.workspace-menu').press('Enter');await page.locator('.workspace-drawer').waitFor();}
    assert.equal(await nav.locator('a').count(),9);assert.equal(await nav.locator('a[aria-current]').getAttribute('href'),'/admin');
    if(width<1000){await page.keyboard.press('Escape');assert.equal(await page.locator('.workspace-menu').evaluate(el=>el===document.activeElement),true);}
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/admin-${locale}-${theme}-${width}.png`});}
   }
   for(const route of ['catalog','requests','teams','taxonomy','moderation','accounts','reports','audit']){const response=await page.goto('/admin/'+route);assert.equal(response.status(),200);for(const width of [320,360,1280]){await page.setViewportSize({width,height:900});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'admin '+route+' fits '+width);}assert.equal(await page.locator('.workspace-sidebar a[aria-current]').getAttribute('href'),'/admin/'+route);}
  }finally{await context.close();}
 }}finally{await browser.close();}
});
