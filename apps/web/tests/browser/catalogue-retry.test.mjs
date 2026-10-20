import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {readFile} from 'node:fs/promises';
import {spawnSync} from 'node:child_process';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Disposable local fixture required');
function sql(input){const result=spawnSync('docker',['compose','--env-file',fixture.env_file,'-p',fixture.project_name,'exec','-T','postgres','psql','-U','qmanga','-d','qmanga','-v','ON_ERROR_STOP=1'],{input,encoding:'utf8',timeout:30000});assert.equal(result.status,0,result.stderr);}
function api(action){const result=spawnSync('docker',['compose','--env-file',fixture.env_file,'-p',fixture.project_name,action,'api'],{encoding:'utf8',timeout:30000});assert.equal(result.status,0,result.stderr);}
test('catalogue retry recovers exact search filters sort and cursor after API outage',{timeout:150000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});let paused=false;const titleId=randomUUID();
 try{
  sql(`INSERT INTO titles(id,slug,name_kk,is_listed) VALUES('${titleId}','outage-${titleId}','Synthetic outage fixture',true)`);
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);
  const source=await context.request.get('/api/v1/titles?q=Synthetic&kind=manga&sort=name&limit=1');assert.equal(source.status(),200);
  const cursor=(await source.json()).next_cursor;assert.ok(cursor,'fixture must have enough public titles to supply a real cursor');
  const exact='/catalog?q='+encodeURIComponent('Synthetic')+'&kind=manga&sort=name&cursor='+encodeURIComponent(cursor);
  const page=await context.newPage();page.setDefaultTimeout(10000);
  await page.goto(exact);await page.locator('.catalogue-content').count();
  api('pause');paused=true;
  await page.reload({timeout:100000});
  const retry=page.getByRole('link',{name:'Повторить',exact:true});await retry.waitFor();
  assert.equal(await retry.getAttribute('href'),exact);
  api('unpause');paused=false;
  await retry.click();await page.getByRole('heading',{name:'Каталог',exact:true}).waitFor();
  await retry.waitFor({state:'detached'});
  assert.equal(new URL(page.url()).pathname+new URL(page.url()).search,exact);
  assert.equal((await context.cookies()).find(cookie=>cookie.name==='qmanga_locale').value,'ru');
  await page.reload();assert.equal(await page.getByRole('link',{name:'Повторить',exact:true}).count(),0);
 }finally{try{if(paused)api('unpause');}finally{sql(`DELETE FROM titles WHERE id='${titleId}'`);await context.close();await browser.close();}}
});
