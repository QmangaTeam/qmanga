import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('project cards display scoped server statistics and covers without a duplicate request strip',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:fixture.app_url});
  try{
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
   const response=await context.request.get('/api/v1/studio/projects');assert.equal(response.status(),200);
   const {items}=await response.json();assert.ok(items.length>0);const item=items.find(item=>item.id===fixture.project_id);assert.ok(item);
   assert.equal(typeof item.chapter_count,'number');assert.equal(typeof item.score,'number');assert.equal(typeof item.cover_url,'string');
   const page=await context.newPage();
   for(const theme of ['dark','light'])for(const width of [320,360,1280]){
    await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);await page.setViewportSize({width,height:900});await page.goto('/studio/projects');
    const card=page.locator('.project-card').filter({has:page.locator(`a.manage[href="/studio/projects/${fixture.project_id}"]`)});await card.waitFor();
    assert.deepEqual(await card.locator('dd').allTextContents(),[String(item.chapter_count),item.rating_count?item.score.toFixed(2):'—']);
    if(item.cover_url)assert.equal(await card.locator('img').getAttribute('src'),item.cover_url);
    assert.equal(await page.locator('a.pending').count(),0);assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/projects-${theme}-${width}.png`});}
   }
   for(const theme of ['dark','light'])for(const width of [320,360,1280]){
    await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
    await page.setViewportSize({width,height:900});await page.goto('/studio/teams');
    await page.locator('section.team').waitFor();
    const overview=page.locator('.team-overview');
    assert.equal(await overview.getByRole('heading',{level:1}).count(),1);
    assert.equal(await overview.getByRole('combobox').count(),1);
    assert.equal(await overview.getByRole('button').count(),1);
    await page.locator('details.management>summary').click();
    assert.ok(await page.locator('.member-row').count()>0);
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    await page.evaluate(()=>scrollTo(0,0));
    if(process.env.QMANGA_SCREENSHOT_DIR)await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/teams-${theme}-${width}.png`});
   }

  }finally{await context.close();}
 }finally{await browser.close();}
});
