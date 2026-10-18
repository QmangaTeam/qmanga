import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {join} from 'node:path';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required; metrics tests never skip.');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname))throw Error('Local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
async function login(admin){const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);const response=await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email:admin?fixture.admin_email:fixture.owner_email,password:admin?fixture.admin_password:fixture.owner_password}});assert.equal(response.status(),200);return context;}
test('private metrics periods render the same server data as a readable table on phone and desktop',{timeout:60000},async()=>{
 const context=await login(true),page=await context.newPage();try{
  const guest=await browser.newContext({baseURL});assert.equal((await guest.request.get('/api/v1/admin/metrics')).status(),401);await guest.close();
  await page.goto('/admin');await page.getByRole('heading',{name:'Обзор сайта',exact:true}).waitFor();
  for(const period of [7,30,90]){
   const received=page.waitForResponse(response=>response.url().endsWith('/admin/metrics?days='+period));await page.getByRole('button',{name:period+' дней',exact:true}).click();const response=await received;assert.equal(response.status(),200);assert.match(response.headers()['cache-control'],/no-store/);const report=await response.json();assert.equal(report.days.length,period);assert.equal(JSON.stringify(report).includes('@'),false);await page.locator('tbody tr').nth(period-1).waitFor();assert.equal(await page.locator('tbody tr').count(),period);const first=page.locator('tbody tr').first();assert.equal(await first.locator('th').innerText(),report.days[0].date);assert.equal(Number(await first.locator('td').nth(2).innerText()),report.days[0].completions);
  }
  const screenshotDir=process.env.QMANGA_SCREENSHOT_DIR;if(screenshotDir)await mkdir(screenshotDir,{recursive:true});
  for(const width of [320,360,1280])for(const theme of ['dark','light']){await page.setViewportSize({width,height:800});await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);if(screenshotDir)await page.screenshot({path:join(screenshotDir,`metrics-${width}-${theme}.png`),fullPage:true});}
  assert.equal((await context.request.get('/api/v1/admin/metrics?days=365')).status(),400);
 }finally{await context.close();}
});
test('translator dashboard uses own membership and cannot obtain administrative data',{timeout:60000},async()=>{const context=await login(false),page=await context.newPage();try{assert.equal((await context.request.get('/api/v1/admin/metrics?days=7')).status(),403);const response=await context.request.get('/api/v1/studio/metrics?days=7');assert.equal(response.status(),200);const report=await response.json();assert.equal(report.days.length,7);assert.equal(report.days.every(day=>day.verified===0&&day.ratings===0),true);assert.equal(JSON.stringify(report).includes('@'),false);await page.goto('/studio/stats');await page.getByRole('heading',{name:'Статистика команд',exact:true}).waitFor();await page.locator('tbody tr').nth(29).waitFor();}finally{await context.close();}});
