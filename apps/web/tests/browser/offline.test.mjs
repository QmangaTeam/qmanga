import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';

if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required; offline acceptance must not be skipped.');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Local fixture required.');

test('cold offline entry reads saved bytes; eviction and retry recover without cloning reactive proxies',async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});
 try{
  let page=await context.newPage();
  await page.goto('/read/'+fixture.release_id);
  await page.locator('.fragment img').first().waitFor();
  await page.locator('.reader-controls').getByRole('button',{name:'Оқу баптаулары',exact:true}).click();
  await page.getByRole('button',{name:/Тарауды құрылғыға сақтау/}).click();
  await page.getByText('Толық сақталды',{exact:false}).waitFor({timeout:30000});
  const quality=await page.evaluate(()=>JSON.parse(localStorage.getItem('qmanga:reader:preferences:v1')).quality);
  await page.close();
  await context.setOffline(true);
  page=await context.newPage();
  await page.goto('/saved');
  await page.getByRole('link',{name:'Желісіз оқу',exact:true}).click();
  await page.locator('.fragment img').first().waitFor();
  await page.locator('.fragment img').first().evaluate(image=>image.decode());
  assert.equal(await page.locator('.fragment img').first().evaluate(image=>image.complete&&image.naturalWidth>0),true);
  assert.equal(await page.evaluate(()=>JSON.parse(localStorage.getItem('qmanga:reader:preferences:v1')).quality),quality);
  await page.evaluate(async()=>{const cache=await caches.open('qmanga-saved-v1');const keys=await cache.keys();if(!keys.length)throw Error('Expected saved image');await cache.delete(keys[0]);});
  await page.goto('/saved');
  await page.getByRole('button',{name:'Жүктеуді жалғастыру',exact:true}).waitFor();
  assert.equal(await page.getByRole('link',{name:'Желісіз оқу',exact:true}).count(),0);
  await context.setOffline(false);
  await page.getByRole('button',{name:'Жүктеуді жалғастыру',exact:true}).click();
  await page.getByRole('link',{name:'Желісіз оқу',exact:true}).waitFor({timeout:30000});
  await page.getByRole('button',{name:'Тексеру',exact:true}).click();
  await page.getByRole('link',{name:'Желісіз оқу',exact:true}).waitFor();
  assert.equal(await page.getByRole('button',{name:'Жүктеуді жалғастыру',exact:true}).count(),0);
 }finally{await context.close();await browser.close();}
});
