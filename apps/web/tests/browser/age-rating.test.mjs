import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {readFile,mkdir} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {join} from 'node:path';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});

test('age ratings persist through admin, public cards and exact URL filters without classifying old titles',{timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800}});
 const created=[];
 let headers;
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}})).status(),200);
  const csrf=(await (await context.request.get('/api/v1/auth/csrf')).json()).csrf_token;
  headers={Origin:fixture.app_url,'X-CSRF-Token':csrf};
  const prefix='AgeFixture-'+randomUUID().slice(0,8);
  for(const age of [0,12,16,18,18]){
   const data={slug:prefix.toLowerCase()+'-'+age+'-'+created.length,name:prefix+' '+age+' '+created.length,original_name:'',aliases:[],authors:[],kind:'manga',original_status:'ongoing',translation_status:'ongoing',description:'Synthetic age classification fixture.',direction:'ltr',genre_ids:[],listed:true,demo:true,age_rating:age};
   const response=await context.request.post('/api/v1/admin/titles',{headers,data});assert.equal(response.status(),201);
   created.push({...await response.json(),data});
  }
  for(const age of [0,12,16,18]){
   const response=await context.request.get('/api/v1/titles?'+new URLSearchParams({q:prefix,age_rating:String(age),limit:'1'}));
   assert.equal(response.status(),200);const page=await response.json();assert.equal(page.items.length,1);assert.equal(page.items[0].age_rating,age);
   if(age===18){
    assert.ok(page.next_cursor);
    const next=await (await context.request.get('/api/v1/titles?'+new URLSearchParams({q:prefix,age_rating:'18',limit:'1',cursor:page.next_cursor}))).json();
    assert.equal(next.items[0].age_rating,18);assert.notEqual(next.items[0].id,page.items[0].id);assert.equal(next.next_cursor,null);
   }
  }
  for(const invalid of ['14','-1','999','nope'])assert.ok([400,422].includes((await context.request.get('/api/v1/titles?age_rating='+invalid)).status()));
  const adult=created[3];
  const legacy={...adult.data};delete legacy.age_rating;
  assert.equal((await context.request.put('/api/v1/admin/titles/'+adult.id,{headers,data:legacy})).status(),200);
  assert.equal((await (await context.request.get('/api/v1/titles/'+adult.slug)).json()).age_rating,18,'old writer cannot erase classification');
  assert.equal((await context.request.put('/api/v1/admin/titles/'+adult.id,{headers,data:{...adult.data,age_rating:14}})).status(),422);
  const page=await context.newPage();
  await page.goto('/admin/catalog?title='+adult.id);
  // The route accepts title selection via the normal picker, independent of any query shortcut.
  const picker=page.locator('select').first();await picker.locator('option').nth(1).waitFor({state:'attached'});
  for(let batch=0;batch<3&&!(await picker.locator(`option[value="${adult.id}"]`).count());batch++){
   const count=await picker.locator('option').count();
   await page.getByRole('button',{name:'Ещё тайтлы',exact:true}).click();
   await picker.locator('option').nth(count).waitFor({state:'attached'});
  }
  await picker.selectOption(adult.id);
  const ageSelect=page.getByRole('combobox',{name:'Возрастная маркировка',exact:true});
  await page.waitForFunction(()=>document.querySelector('select[name="age_rating"]')?.value==='18');
  assert.equal(await ageSelect.inputValue(),'18');
  await ageSelect.selectOption('16');await page.getByRole('button',{name:'Сохранить тайтл',exact:true}).click();
  await page.getByText('Сохранено',{exact:true}).waitFor();
  assert.equal((await (await context.request.get('/api/v1/titles/'+adult.slug)).json()).age_rating,16);
  await page.goto('/catalog?'+new URLSearchParams({q:prefix,age_rating:'18'}));
  await page.getByRole('button',{name:'Фильтры',exact:true}).click();
  const filtersDialog=page.getByRole('dialog',{name:'Фильтры каталога',exact:true});
  assert.equal(await filtersDialog.getByRole('combobox',{name:'Возрастная маркировка',exact:true}).inputValue(),'18');
  assert.equal(await page.locator('.cover-grid .cover-card').count(),1);
  assert.equal(await page.locator('.cover-card .age-rating.adult').count(),1);
  await filtersDialog.getByRole('combobox',{name:'Возрастная маркировка',exact:true}).selectOption('12');
  await filtersDialog.getByRole('button',{name:'Применить',exact:true}).click();
  await page.waitForURL(url=>url.pathname==='/catalog'&&url.searchParams.get('age_rating')==='12');
  assert.equal(new URL(page.url()).searchParams.get('q'),prefix);assert.equal(new URL(page.url()).searchParams.has('cursor'),false);
  await page.locator('.cover-card .age-rating').filter({hasText:'12+'}).waitFor();
  await page.goto('/titles/'+created[0].slug);assert.equal(await page.locator('.age-rating').count(),0);
  for(const locale of ['kk','ru'])for(const theme of ['light','dark']){
   await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);
   await page.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   await page.goto('/titles/'+created[4].slug);
   const badge=page.locator('.age-rating.adult');await badge.waitFor();
   assert.equal(await badge.innerText(),'18+');
   const rgb=await badge.evaluate(el=>getComputedStyle(el).backgroundColor.match(/\d+/g).map(Number));
   assert.ok(rgb[0]>rgb[1]*2&&rgb[0]>rgb[2]*2,'18+ uses a red surface in both themes');
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:800});
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,locale+' '+theme+' '+width);
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`age-${locale}-${theme}-${width}.png`)});}
   }
  }
  const reader=await browser.newContext({baseURL:fixture.app_url});
  try{
   await reader.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});
   const token=(await (await reader.request.get('/api/v1/auth/csrf')).json()).csrf_token;
   assert.equal((await reader.request.put('/api/v1/admin/titles/'+adult.id,{headers:{Origin:fixture.app_url,'X-CSRF-Token':token},data:{...adult.data,age_rating:12}})).status(),403);
  }finally{await reader.close();}
 }finally{
  for(const item of created)if(headers){
   const response=await context.request.put('/api/v1/admin/titles/'+item.id,{headers,data:{...item.data,listed:false}});
   assert.equal(response.status(),200,'hide only this test\'s synthetic announcement');
  }
  await context.close();
 }
});
