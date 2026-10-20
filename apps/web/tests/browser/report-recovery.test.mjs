import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';

if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Disposable local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
async function session(role='owner',locale='ru'){
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800},reducedMotion:'reduce'});
 await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);
 const result=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture[role+'_email'],password:fixture[role+'_password']}});
 assert.equal(result.status(),200,'fixture login');
 return context;
}

test('report queue failure never claims empty and can retry without a reload',{timeout:30000},async()=>{
 const context=await session('admin');
 try{
  const page=await context.newPage();let failed=true;
  await page.route('**/api/v1/moderation/reports?*',route=>failed?route.abort():route.fulfill({json:{items:[],next:'',has_more:false}}));
  await page.goto('/moderation/reports');await page.getByText('Подключись к сети и попробуй снова.',{exact:true}).waitFor();
  assert.equal(await page.getByText('Необработанных жалоб нет.',{exact:true}).count(),0,'network failure is not an empty queue');
  failed=false;await page.getByRole('button',{name:'Повторить',exact:true}).click();await page.getByText('Необработанных жалоб нет.',{exact:true}).waitFor();
  assert.equal(await page.getByText('Подключись к сети и попробуй снова.',{exact:true}).count(),0,'retry clears stale error');
 }finally{await context.close();}
});

