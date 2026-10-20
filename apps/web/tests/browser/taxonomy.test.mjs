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
test('taxonomy admin CRUD, title assignment and public combined URL filters',{timeout:90000},async()=>{
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:800},reducedMotion:'reduce'});
 const suffix=randomUUID().slice(0,8),slug='taxonomy-'+suffix,name='Taxonomy '+suffix;
 let headers,category,genre,collection,title,titleData;
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.admin_email,password:fixture.admin_password}})).status(),200);
  headers={Origin:fixture.app_url,'X-CSRF-Token':(await (await context.request.get('/api/v1/auth/csrf')).json()).csrf_token};
  const page=await context.newPage();page.setDefaultTimeout(10000);
  await page.goto('/admin/taxonomy');
  const taxonomyForm=page.locator('section.panel').first().locator('form');
  await taxonomyForm.getByRole('combobox',{name:'Тип',exact:true}).selectOption('categories');
  await taxonomyForm.getByRole('textbox',{name:'Название',exact:true}).fill(name);
  await taxonomyForm.getByRole('textbox',{name:'Имя в ссылке',exact:true}).fill(slug);
  await taxonomyForm.getByRole('button',{name:'Сохранить',exact:true}).click();
  await page.getByRole('status').filter({hasText:'Сохранено'}).waitFor();
  category=(await (await context.request.get('/api/v1/admin/taxonomy')).json()).categories.find(item=>item.slug===slug);assert.ok(category);
  await taxonomyForm.getByRole('combobox',{name:'Выбрать',exact:true}).selectOption(category.id);
  await taxonomyForm.getByRole('textbox',{name:'Название',exact:true}).fill(name+' revised');
  await taxonomyForm.getByRole('button',{name:'Сохранить',exact:true}).click();
  await page.waitForFunction(({id,value})=>Array.from(document.querySelectorAll('option')).some(option=>option.value===id&&option.textContent.startsWith(value+' ·')),{id:category.id,value:name+' revised'});
  let response=await context.request.post('/api/v1/admin/genres',{headers,data:{name,slug}});assert.equal(response.status(),201);genre=await response.json();
  titleData={name,slug,kind:'manga',original_status:'ongoing',translation_status:'ongoing',description:'Synthetic taxonomy fixture',direction:'ltr',aliases:[],authors:[],genre_ids:[genre.id],listed:true,demo:true,age_rating:18};
  response=await context.request.post('/api/v1/admin/titles',{headers,data:titleData});assert.equal(response.status(),201);title=await response.json();
  await page.goto('/admin/catalog?title='+title.id);
  await page.waitForFunction(id=>document.querySelector('section.panel select')?.value===id,title.id);
  const checkbox=page.getByRole('checkbox',{name:name+' revised',exact:true});await checkbox.check();
  await page.getByRole('button',{name:'Сохранить тайтл',exact:true}).click();await page.getByRole('status').filter({hasText:'Сохранено'}).waitFor();
  response=await context.request.post('/api/v1/admin/collections',{headers,data:{name,slug,description:'Editorial collection',position:1,visible:true,title_ids:[title.id]}});assert.equal(response.status(),200);collection=await response.json();
  for(const locale of ['kk','ru'])for(const theme of ['light','dark']){
   await context.addCookies([{name:'qmanga_locale',value:locale,url:fixture.app_url}]);await page.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:800});
    await page.goto('/catalog?'+new URLSearchParams({category:slug,collection:slug,genre:slug,age_rating:'18'}));
    await page.locator('.cover-grid .cover-card').waitFor();assert.equal(await page.locator('.cover-grid .cover-card').count(),1);
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);
    const triggerName=locale==='ru'?'Фильтры':'Сүзгілер';
    const mobile=width<1024;
    if(mobile)await page.getByRole('button',{name:triggerName,exact:true}).click();
    const fields=mobile?page.getByRole('dialog',{name:triggerName,exact:true}):page.locator('aside.catalogue-filters');
    assert.equal(await fields.locator('input[type="checkbox"][value="18"]').isChecked(),true);
    const collectionSection=fields.locator('section.filter-section').filter({hasText:locale==='ru'?'Редакционная подборка':'Редакциялық жинақ'});
    const collectionBox=collectionSection.locator(`input[type="checkbox"][value="${slug}"]`);
    assert.equal(await collectionBox.isChecked(),true);
    await fields.getByRole('searchbox',{name:locale==='ru'?'Поиск тегов':'Тег іздеу',exact:true}).fill(name);
    const tag=fields.locator('.tag-browser').getByRole('checkbox',{name,exact:true});
    await tag.waitFor();
    assert.equal(await tag.isChecked(),true);
    assert.equal(await fields.getByRole('button',{name:locale==='ru'?'Применить':'Қолдану',exact:true}).count(),0);
    assert.ok((await page.locator('.active-filters').innerText()).includes(name+' revised'),'the category chip uses the renamed taxonomy');
    await collectionBox.uncheck();
    await page.waitForURL(url=>!url.searchParams.has('collection'));
    assert.equal(await tag.isChecked(),true,'removing one filter keeps the selected tag');
    assert.equal(new URL(page.url()).searchParams.get('genre'),slug);
    assert.equal(new URL(page.url()).searchParams.get('category'),slug);
    assert.equal(new URL(page.url()).searchParams.get('age_rating'),'18');
    if(mobile){await page.keyboard.press('Escape');await fields.waitFor({state:'hidden'});}
   }
  }
  await page.goto('/titles/'+slug);await page.locator('.title-genres a[href="/catalog?category='+slug+'"]').waitFor();await page.locator('.title-genres a[href="/catalog?collection='+slug+'"]').waitFor();
  const anonymous=await browser.newContext({baseURL:fixture.app_url});try{assert.equal((await anonymous.request.get('/api/v1/admin/taxonomy')).status(),401);}finally{await anonymous.close();}
  response=await context.request.put('/api/v1/admin/collections/'+collection.id,{headers,data:{...collection,visible:false,title_ids:[title.id]}});assert.equal(response.status(),200);
  assert.equal((await (await context.request.get('/api/v1/titles?collection='+slug)).json()).items.length,0);
  assert.ok(!(await (await context.request.get('/api/v1/taxonomy')).json()).collections.some(item=>item.id===collection.id));
 }finally{
  if(headers){for(const [kind,item] of [['collections',collection],['categories',category],['genres',genre]])if(item)assert.equal((await context.request.delete('/api/v1/admin/'+kind+'/'+item.id,{headers})).status(),204);if(title)assert.equal((await context.request.put('/api/v1/admin/titles/'+title.id,{headers,data:{...titleData,genre_ids:[],listed:false}})).status(),200);}
  await context.close();
 }
});
