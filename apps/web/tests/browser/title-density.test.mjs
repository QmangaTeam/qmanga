import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('metadata and contributors grow naturally until their viewport bounds',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const theme of ['dark','light']){const context=await browser.newContext({baseURL:fixture.app_url,colorScheme:theme,reducedMotion:'reduce'});
  try{const page=await context.newPage();await page.goto('/titles/'+fixture.title_slug+'?tab=about');await page.waitForLoadState('networkidle');
   for(const width of [320,360,730,960,1440]){await page.setViewportSize({width,height:807});
    for(const selector of ['.metadata-rail','.contributors'])if(await page.locator(selector).count()){
     const metrics=await page.locator(selector).evaluate(e=>({height:e.clientHeight,content:e.scrollHeight,cap:parseFloat(getComputedStyle(e).maxHeight),overflow:getComputedStyle(e).overflowY}));
     assert.ok(metrics.height<=metrics.cap+1,'dense information stays within its cap');assert.equal(metrics.overflow,'auto');
     if(metrics.content<=metrics.cap)assert.ok(Math.abs(metrics.height-metrics.content)<2,'short information does not get a forced empty height');
    }
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
   }
  }finally{await context.close()}
 }}finally{await browser.close()}
});
