import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('title choices preserve scroll and animate only without reduced motion',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const width of [360,688,1280])for(const reducedMotion of ['reduce','no-preference']){
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width,height:807},reducedMotion});
  try{const page=await context.newPage();await page.goto('/titles/'+fixture.title_slug+'?tab=about');await page.waitForLoadState('networkidle');
   await page.evaluate(()=>scrollTo(0,document.querySelector('.title-tabs').getBoundingClientRect().top+scrollY-120));
   const before=await page.evaluate(()=>scrollY);
   await page.locator('.title-tabs a[href*="tab=chapters"]').click();await page.waitForURL(url=>url.searchParams.get('tab')==='chapters');await page.locator('.chapter-search').waitFor();
   assert.ok(Math.abs(await page.evaluate(()=>scrollY)-before)<3,'tab choice must not jump to content/footer');assert.equal(new URL(page.url()).hash,'');
   const duration=await page.locator('.title-tabs').evaluate(e=>getComputedStyle(e,'::before').transitionDuration);assert.equal(duration,reducedMotion==='reduce'?'0s':'0.18s');
   const animation=await page.locator('.tab-panel').evaluate(e=>getComputedStyle(e).animationName);if(reducedMotion==='reduce')assert.equal(animation,'none');else assert.ok(animation.endsWith('tab-enter'));
   await page.goBack();await page.waitForURL(url=>url.searchParams.get('tab')==='about');await page.locator('.title-tabs a[href*="tab=about"][aria-current]').waitFor();assert.equal(await page.locator('.title-tabs a[aria-current]').getAttribute('href'),'/titles/'+fixture.title_slug+'?tab=about');
  }finally{await context.close()}
 }}finally{await browser.close()}
});
