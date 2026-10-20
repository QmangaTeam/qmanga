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

test('account load more uses the applied search until another search is submitted',{timeout:30000},async()=>{
 const context=await session('admin');
 try{
  const page=await context.newPage();const calls=[];
  await page.route('**/api/v1/moderation/accounts?*',async route=>{
   const url=new URL(route.request().url());const query=url.searchParams.get('q')||'',after=url.searchParams.get('after')||'';calls.push({query,after});
   const id=after?'second':'first';await route.fulfill({json:{items:[{id,name:'Synthetic '+id,role:'reader',blocked:false,version:1}],has_more:!after}});
  });
  await page.goto('/admin/accounts');await page.getByText('Synthetic first',{exact:true}).waitFor();
  const query=page.getByRole('textbox',{name:'Имя, ID или точная почта'});await query.fill('new-unsubmitted-search');
  await page.getByRole('button',{name:'Загрузить ещё',exact:true}).click();await page.getByText('Synthetic second',{exact:true}).waitFor();
  assert.deepEqual(calls.at(-1),{query:'',after:'first'},'continuation stays bound to applied criteria');
  await page.getByRole('button',{name:'Найти',exact:true}).click();await page.getByText('Synthetic second',{exact:true}).waitFor({state:'hidden'});
  assert.deepEqual(calls.at(-1),{query:'new-unsubmitted-search',after:''});assert.equal(await page.getByText('Synthetic second',{exact:true}).count(),0);
 }finally{await context.close();}
});
