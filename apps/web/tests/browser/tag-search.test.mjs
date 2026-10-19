import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {chromium} from 'playwright';
import {readFile} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw Error('Local fixture required');
let browser;
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{await browser?.close();});
test('tag search finds a tag by its name, applies it and filters the catalogue',{timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800},reducedMotion:'reduce'});
 const suffix=randomUUID().slice(0,8),slug='tag-'+suffix,name='Тег '+suffix;
 let headers,genre,title,titleData;
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}})).status(),200);
  headers={Origin:fixture.app_url,'X-CSRF-Token':(await (await context.request.get('/api/v1/auth/csrf')).json()).csrf_token};
  let response=await context.request.post('/api/v1/admin/genres',{headers,data:{name,slug}});assert.equal(response.status(),201);genre=await response.json();
  titleData={name,slug:'tagged-'+suffix,kind:'manga',original_status:'ongoing',translation_status:'ongoing',description:'Synthetic tag fixture',direction:'ltr',aliases:[],authors:[],genre_ids:[genre.id],listed:true,demo:true,age_rating:12};
  response=await context.request.post('/api/v1/admin/titles',{headers,data:titleData});assert.equal(response.status(),201);title=await response.json();
  const page=await context.newPage();page.setDefaultTimeout(10000);
  await page.goto('/catalog');
  await page.getByRole('button',{name:'Фильтры',exact:true}).click();
  const dialog=page.getByRole('dialog',{name:'Фильтры каталога',exact:true});await dialog.waitFor();
  await dialog.getByRole('combobox',{name:'Теги',exact:true}).fill(name);
  const option=dialog.locator('.tag-search-results').getByRole('option',{name,exact:true});
  await option.waitFor();
  await option.click();
  assert.equal(await dialog.locator('.tag-search-selected button').count(),1,'a picked tag becomes a chip');
  assert.equal(await dialog.locator('input[type="hidden"][name="genre"]').inputValue(),slug);
  await dialog.getByRole('button',{name:'Применить',exact:true}).click();
  await page.waitForURL(url=>url.searchParams.get('genre')===slug);
  await page.locator('.cover-grid .cover-card').waitFor();
  assert.equal(await page.locator('.cover-grid .cover-card').count(),1,'the tag filter keeps the tagged title');
  assert.equal(await page.locator('.active-filters span').first().textContent(),name);
  response=await context.request.get('/api/v1/titles?genre='+slug);assert.equal((await response.json()).items.length,1);
  response=await context.request.get('/api/v1/titles?q='+encodeURIComponent(name));assert.equal((await response.json()).items.length,1,'the free text search finds the title by tag name');
 }finally{
  if(headers){if(title)await context.request.put('/api/v1/admin/titles/'+title.id,{headers,data:{...titleData,genre_ids:[],listed:false}});if(genre)await context.request.delete('/api/v1/admin/genres/'+genre.id,{headers});}
  await context.close();
 }
});
