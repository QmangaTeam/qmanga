import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('compact title header reaches reading early and fits long names across screens',{timeout:60000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:844},colorScheme:theme});
  try{await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);const page=await context.newPage();await page.goto('/titles/'+fixture.title_slug+'?tab=about');await page.waitForLoadState('networkidle');
   for(const width of [320,360,768,1024,1440]){
    await page.setViewportSize({width,height:900});
    const m=await page.evaluate(()=>{const cover=document.querySelector('.cover-wrap').getBoundingClientRect(),heading=document.querySelector('.title-heading').getBoundingClientRect(),read=document.querySelector('.read-button').getBoundingClientRect(),banner=document.querySelector('.title-banner');return {coverRight:cover.right,headingLeft:heading.left,readTop:read.top,bannerHeight:banner?banner.getBoundingClientRect().height:0,overflow:document.documentElement.scrollWidth>innerWidth};});
    assert.equal(m.overflow,false,`${locale}/${theme}/${width} fits`);
    if(width<768){assert.ok(m.coverRight<=m.headingLeft+1,'mobile cover beside heading');assert.ok(m.readTop<520,'reading action appears in first screen');}
    if(width>=960){assert.ok(m.bannerHeight<=180,'desktop banner stays shallow');assert.ok(m.readTop<620,'desktop reading action stays near header');}
   }
   await page.setViewportSize({width:320,height:844});await page.locator('h1').evaluate(e=>e.textContent='Өте ұзақ атауы бар шығарма — '+ 'ҰзақАтау'.repeat(12));
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false,'long title reflows');
  }finally{await context.close();}
 }}finally{await browser.close();}
});
