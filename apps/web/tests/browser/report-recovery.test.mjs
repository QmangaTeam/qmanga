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


test('report retry retains loaded rows and repeats the failed continuation cursor',{timeout:30000},async()=>{
 const context=await session('admin');
 try{
  const page=await context.newPage();let failNext=true;const calls=[];
  await page.route('**/api/v1/moderation/reports?*',route=>{
   const after=new URL(route.request().url()).searchParams.get('after');calls.push(after);
   if(after&&failNext)return route.abort();
   return route.fulfill({json:{items:after?[]:[{id:'synthetic-report',title_name:'Synthetic report',target_type:'title',target_id:'synthetic-title',category:'other',reason:'Synthetic report reason',created_at:'2026-10-20T10:00:00Z',history:[]}],next:after?'':'next-report',has_more:!after}});
  });
  await page.goto('/moderation/reports');await page.getByRole('heading',{name:'Synthetic report',exact:true}).waitFor();
  await page.getByRole('button',{name:'Загрузить ещё',exact:true}).click();await page.getByRole('alert').waitFor();
  assert.equal(await page.getByRole('heading',{name:'Synthetic report',exact:true}).count(),1);
  failNext=false;await page.getByRole('button',{name:'Повторить',exact:true}).click();await page.getByRole('alert').waitFor({state:'hidden'});await page.getByRole('button',{name:'Загрузить ещё',exact:true}).waitFor({state:'hidden'});
  assert.deepEqual(calls,['','next-report','next-report']);assert.equal(await page.getByRole('heading',{name:'Synthetic report',exact:true}).count(),1);assert.equal(await page.getByRole('button',{name:'Загрузить ещё',exact:true}).count(),0);
 }finally{await context.close();}
});
