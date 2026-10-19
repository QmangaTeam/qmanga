import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('status updates the title button immediately and ratings save from ten accessible choices',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
   const page=await context.newPage();page.setDefaultTimeout(7000);await page.goto('/titles/'+(fixture.title_slug||fixture.title_id));
   await page.locator('.library-trigger').click();const library=page.locator('.action-dialog').filter({has:page.locator('.selection')});
   await library.getByRole('button',{name:'Брошено',exact:true}).click();
   await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Брошено'));
   const boxes=await library.locator('.status-choices button').evaluateAll(elements=>elements.map(e=>e.getBoundingClientRect()));
   assert.ok(boxes.every(b=>Math.abs(b.x-boxes[0].x)<1),'statuses form one column');
   await library.getByRole('button',{name:'Закрыть',exact:true}).click();
   await page.reload();await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Брошено'));
   await page.locator('.rating-trigger').click();const dialog=page.locator('.rating-dialog');
   const choices=dialog.locator('.rating-choices button');assert.equal(await choices.count(),10);
   assert.equal(await dialog.getByRole('combobox').count(),0);assert.equal(await dialog.getByRole('button',{name:'Сохранить',exact:true}).count(),0);
   await choices.nth(6).click();await dialog.waitFor({state:'hidden'});
   assert.equal(await page.locator('.rating-trigger .rate-label').textContent(),'Ваша: 7');
   await page.locator('.rating-trigger').click();assert.equal(await choices.nth(6).getAttribute('aria-pressed'),'true');
   await page.route('**/api/v1/community/titles/*/rating',async route=>route.request().method()==='PUT'?route.fulfill({status:500,json:{error:{code:'internal'}}}):route.continue());
   await choices.nth(7).click();await dialog.getByRole('alert').waitFor();assert.equal(await dialog.isVisible(),true);assert.equal(await choices.nth(6).getAttribute('aria-pressed'),'true');
   await page.unroute('**/api/v1/community/titles/*/rating');
   for(const theme of ['dark','light'])for(const width of [320,360,1280]){
    await page.evaluate(value=>document.documentElement.dataset.theme=value,theme);await page.setViewportSize({width,height:850});
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    const size=await choices.first().boundingBox();assert.ok(size.width>=44&&size.height>=44);
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/rating-${theme}-${width}.png`});}
   }
  }finally{await context.close();}
 }finally{await browser.close();}
});
