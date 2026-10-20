import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {join,dirname} from 'node:path';
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

test('metadata taxonomy stays compact with 720 options and retains ID selections across search',{timeout:60000},async()=>{
 for(const locale of ['ru','kk']){
 const context=await session('owner',locale);
 try{
  const page=await context.newPage();
  await page.route('**/api/v1/taxonomy',async route=>{const response=await route.fetch();const data=await response.json();data.genres=[...data.genres,...Array.from({length:720},(_,i)=>({id:'synthetic-genre-'+i,slug:'synthetic-genre-'+i,name:'Сынақ жанры '+i,name_ru:'Тестовый жанр '+i}))];await route.fulfill({response,json:data});});
  await page.route('**/api/v1/studio/projects/*/title-change-context',async route=>{const response=await route.fetch();const data=await response.json();data.snapshot.data.genre_ids=Array.from({length:13},(_,i)=>'synthetic-genre-'+i);await route.fulfill({response,json:data});});
  await page.goto('/studio/projects/'+fixture.project_id+'/metadata');
  const submit=page.getByRole('button',{name:locale==='ru'?'Отправить на модерацию':'Модерацияға жіберу',exact:true});await submit.waitFor();
  assert.ok(await page.locator('main').evaluate(el=>el.scrollHeight)<6500,'collapsed taxonomy must not force hundreds of rows before submit');
  const picker=page.getByRole('group',{name:locale==='ru'?'Жанры':'Жанрлар',exact:true});
  await picker.getByRole('button',{name:locale==='ru'?'Выбрать':'Таңдау',exact:true}).click();
  const search=picker.getByRole('searchbox');await search.fill(locale==='ru'?'Тестовый жанр 719':'Сынақ жанры 719');
  const choice=picker.getByRole('checkbox',{name:locale==='ru'?'Тестовый жанр 719':'Сынақ жанры 719',exact:true});
  await choice.focus();await page.keyboard.press('Space');assert.equal(await choice.isChecked(),true);
  await search.fill('no-match-for-synthetic-test');assert.equal(await picker.getByRole('checkbox').count(),0);
  await search.fill('719');assert.equal(await choice.isChecked(),true);
  await picker.getByRole('button',{name:locale==='ru'?'Готово':'Дайын',exact:true}).click();
  assert.equal(await picker.getByRole('searchbox').isVisible(),false);
  for(const theme of ['dark','light'])for(const width of [320,360,1280]){await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);await page.setViewportSize({width,height:850});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true,locale+' '+theme+' '+width);}
  await page.getByRole('textbox',{name:locale==='ru'?'Что и почему нужно изменить?':'Нені және не үшін өзгертесіз?',exact:true}).fill('Synthetic taxonomy retention regression');
  await page.getByRole('textbox',{name:locale==='ru'?'Ссылки на источники, по одной на строке':'Дереккөз сілтемелері, әр жолға біреуі',exact:true}).fill('https://example.org/synthetic');
  const sent=new Promise(resolve=>page.route('**/api/v1/studio/projects/*/title-changes',async route=>{if(route.request().method()!=='POST')return route.continue();resolve(route.request().postData());await route.fulfill({status:503,json:{error:{code:'unavailable'}}});}));
  await submit.click();const payload=await sent;assert.match(payload,/synthetic-genre-719/,'ID selection survives collapsing and search');for(let i=0;i<13;i++)assert.ok(payload.includes('\"synthetic-genre-'+i+'\"'),'existing selection survives beyond twelve IDs');
 }finally{await context.close();}
 }
});

test('refreshing metadata after a failed submit asks before discarding text and evidence',{timeout:45000},async()=>{
 const context=await session();
 try{
  const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id+'/metadata');
  const reason=page.getByRole('textbox',{name:'Что и почему нужно изменить?',exact:true});await reason.fill('Synthetic unsaved explanation');
  const evidence=page.getByLabel('Изображения доказательств',{exact:false});await evidence.setInputFiles(join(dirname(process.env.QMANGA_FIXTURE_FILE),'pages','page-01.png'));
  await page.route('**/api/v1/studio/projects/*/title-changes',async route=>route.request().method()==='POST'?route.fulfill({status:503,json:{error:{code:'unavailable'}}}):route.continue());
  await page.getByRole('button',{name:'Отправить на модерацию',exact:true}).click();await page.getByRole('alert').waitFor();
  let confirmations=0;const dismiss=async dialog=>{confirmations++;await dialog.dismiss();};page.on('dialog',dismiss);
  await page.getByRole('button',{name:'Обновить данные',exact:true}).click();await reason.waitFor();
  assert.equal(confirmations,1,'refresh must ask before discarding a dirty proposal');assert.equal(await reason.inputValue(),'Synthetic unsaved explanation');assert.equal(await evidence.evaluate(el=>el.files.length),1);
  page.off('dialog',dismiss);page.once('dialog',dialog=>dialog.accept());await page.getByRole('button',{name:'Обновить данные',exact:true}).click();await reason.waitFor();assert.equal(await reason.inputValue(),'');assert.equal(await evidence.evaluate(el=>el.files.length),0);
 }finally{await context.close();}
});

