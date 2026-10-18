import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw Error('QMANGA_FIXTURE_FILE required; reading suggestion tests never skip');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Disposable local fixture required');
const user=randomUUID(),other=randomUUID(),title=randomUUID(),slug='suggestion-'+title,teams=[randomUUID(),randomUUID()],projects=[randomUUID(),randomUUID()];
const chapters=Array.from({length:3},()=>({chapter:randomUUID(),release:randomUUID(),revision:randomUUID(),pages:[randomUUID(),randomUUID()]}));
const alternate={release:randomUUID(),revision:randomUUID(),page:randomUUID()};
const email='reading-'+user+'@example.test',otherEmail='reading-'+other+'@example.test';let browser,source;
function sql(input){const result=spawnSync('docker',['compose','--env-file',fixture.env_file,'-p',fixture.project_name,'exec','-T','postgres','psql','-U','qmanga','-d','qmanga','-v','ON_ERROR_STOP=1'],{input,encoding:'utf8',timeout:30000});assert.equal(result.status,0,result.stderr);}
function clonePage(page,revision,position){return `INSERT INTO pages(id,revision_id,position,width,height,status) SELECT '${page}','${revision}',${position},width,height,'ready' FROM pages WHERE id='${source}'; INSERT INTO page_variants(id,page_id,tier,object_key,width,height,byte_size,offset_y,checksum,mime) SELECT md5('${page}'||id)::uuid::text,'${page}',tier,object_key,width,height,byte_size,offset_y,checksum,mime FROM page_variants WHERE page_id='${source}';`;}
before(async()=>{
 browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL});const manifest=await(await context.request.get('/api/v1/releases/'+fixture.release_id+'/manifest')).json();source=manifest.pages[0].id;assert.match(source,/^[a-f0-9-]{36}$/);await context.close();
 let setup=`BEGIN;INSERT INTO users(id,email,password_hash,is_verified,display_name) SELECT '${user}','${email}',password_hash,true,'Suggestion reader' FROM users WHERE email='${fixture.owner_email}'; INSERT INTO users(id,email,password_hash,is_verified,display_name) SELECT '${other}','${otherEmail}',password_hash,true,'Other reader' FROM users WHERE id='${user}'; INSERT INTO titles(id,slug,name_kk,type,is_demo) VALUES('${title}','${slug}','Оқу ұсынысы · Сохранение после чтения','manhwa',true);`;
 for(let i=0;i<2;i++)setup+=`INSERT INTO teams(id,name,slug,is_approved) VALUES('${teams[i]}','Suggestion team ${i}','suggestion-${teams[i]}',true);INSERT INTO projects(id,title_id,team_id) VALUES('${projects[i]}','${title}','${teams[i]}');`;
 for(const [index,ch]of chapters.entries()){setup+=`INSERT INTO chapters(id,title_id,label,order_num) VALUES('${ch.chapter}','${title}','${index+1}',${index+1});INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('${ch.release}','${ch.chapter}','${projects[0]}','${teams[0]}','published','${ch.revision}');INSERT INTO revisions(id,release_id,status) VALUES('${ch.revision}','${ch.release}','published');`;ch.pages.forEach((page,index)=>setup+=clonePage(page,ch.revision,index+1));}
 setup+=`INSERT INTO releases(id,chapter_id,project_id,team_id,status,current_revision_id) VALUES('${alternate.release}','${chapters[0].chapter}','${projects[1]}','${teams[1]}','published','${alternate.revision}');INSERT INTO revisions(id,release_id,status) VALUES('${alternate.revision}','${alternate.release}','published');`+clonePage(alternate.page,alternate.revision,1);sql(setup+'COMMIT;');
});
after(async()=>{try{sql(`BEGIN;DELETE FROM users WHERE id IN('${user}','${other}');DELETE FROM titles WHERE id='${title}';DELETE FROM teams WHERE id IN('${teams[0]}','${teams[1]}');COMMIT;`);}finally{await browser?.close();}});
async function contextFor(account=email){const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);await context.addInitScript(()=>localStorage.setItem('qmanga:reader:preferences:v1',JSON.stringify({mode:'strip',quality:'eco',theme:'dark'})));const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email:account,password:fixture.owner_password}});assert.equal(login.status(),200);return context;}
async function headers(context,id=user){const csrf=(await(await context.request.get('/api/v1/auth/csrf')).json()).csrf_token;return {Origin:baseURL,'X-CSRF-Token':csrf,'X-Qmanga-Account':id};}
async function complete(context,release,revision,page,sequence,baseVersion){const body={event_id:randomUUID(),device_id:'suggestion-browser',sequence,base_version:baseVersion,release_id:release,revision_id:revision,page_id:page,fraction:1,completed:true};const result=await context.request.post('/api/v1/progress/'+title,{headers:await headers(context),data:body});assert.equal(result.status(),200);const accepted=await result.json();assert.equal(accepted.outcome,'accepted');return {body,accepted};}
const prompt=page=>page.getByRole('complementary',{name:'Добавить в библиотеку',exact:true});
async function finishReader(page){await page.locator('.reading-stage').waitFor();const accepted=page.waitForResponse(response=>response.request().method()==='POST'&&response.url().endsWith('/progress/'+title)&&response.request().postDataJSON()?.completed===true);await page.evaluate(()=>scrollTo(0,document.documentElement.scrollHeight));assert.equal((await accepted).status(),200);}

test('preferred team comes from actual translation, remains editable, and third distinct completion prompts once per tab session',{timeout:120000},async()=>{
 const context=await contextFor(),page=await context.newPage();try{
  const first=await complete(context,alternate.release,alternate.revision,alternate.page,1,0);
  const replay=await context.request.post('/api/v1/progress/'+title,{headers:await headers(context),data:first.body});assert.equal((await replay.json()).outcome,'duplicate');
  assert.equal((await(await context.request.get('/api/v1/reading/suggestion/'+title,{headers:await headers(context)})).json()).eligible,false);
  await page.goto('/titles/'+slug);const selection=page.locator('.selection');await selection.getByRole('button',{name:'Читаю',exact:true}).click();await page.waitForFunction(()=>document.querySelector('.selection .preferences select')?.value!== '');
  let stored=(await(await context.request.get('/api/v1/library/'+title,{headers:await headers(context)})).json()).entry;assert.equal(stored.preferred_team_id,teams[1]);
  await selection.getByText('Дополнительные настройки',{exact:true}).click();await selection.getByRole('combobox',{name:'Предпочитаемая команда',exact:true}).selectOption(teams[0]);const saved=page.waitForResponse(r=>r.request().method()==='PUT'&&r.url().endsWith('/library/'+title));await selection.getByRole('button',{name:'Сохранить',exact:true}).click();assert.equal((await saved).status(),204);
  stored=(await(await context.request.get('/api/v1/library/'+title,{headers:await headers(context)})).json()).entry;assert.equal(stored.preferred_team_id,teams[0]);
  assert.equal((await context.request.delete('/api/v1/library/'+title,{headers:await headers(context)})).status(),204);
  await complete(context,chapters[1].release,chapters[1].revision,chapters[1].pages[1],2,1);
  assert.equal((await(await context.request.get('/api/v1/reading/suggestion/'+title,{headers:await headers(context)})).json()).eligible,false);
  let suggestions=0;page.on('request',request=>{if(request.url().includes('/reading/suggestion/'))suggestions++;});
  await page.goto('/read/'+chapters[2].release);
  const middle=await page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/progress/'+title)&&r.request().postDataJSON()?.completed===false);assert.equal(middle.status(),200);assert.equal(suggestions,0,'ordinary progress cannot check before third chapter completes');assert.equal(await prompt(page).count(),0);
  await finishReader(page);await prompt(page).waitFor();assert.equal(suggestions,1);assert.equal((await(await context.request.get('/api/v1/reading/suggestion/'+title,{headers:await headers(context)})).json()).eligible,true);
  await prompt(page).getByRole('button',{name:'Не сейчас',exact:true}).click();assert.equal(await prompt(page).count(),0);
  const before=suggestions;await page.reload();await finishReader(page);assert.equal(suggestions,before,'same tab session dismissal survives reload');assert.equal(await prompt(page).count(),0);
 }finally{await context.close();}
});

test('an actual login in another tab closes the first account prompt; guests never receive it',{timeout:90000},async()=>{
 const context=await contextFor(),page=await context.newPage();try{
  await page.goto('/read/'+chapters[2].release);await finishReader(page);await prompt(page).waitFor();
  const second=await context.newPage();await second.goto('/login');await second.locator('input[type=email]').fill(otherEmail);await second.locator('input[type=password]').fill(fixture.owner_password);await second.locator('button[type=submit]').click();await second.waitForURL(url=>!url.pathname.startsWith('/login'));await prompt(page).waitFor({state:'detached'});assert.equal(await page.locator('.library-prompt').count(),0);
 }finally{await context.close();}
 const guest=await browser.newContext({baseURL,viewport:{width:360,height:800}});try{await guest.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);await guest.addInitScript(()=>localStorage.setItem('qmanga:reader:preferences:v1',JSON.stringify({mode:'strip',quality:'eco'})));const page=await guest.newPage();let suggestions=0;page.on('request',r=>{if(r.url().includes('/reading/suggestion/'))suggestions++;});await page.goto('/read/'+chapters[2].release);await page.evaluate(()=>scrollTo(0,document.documentElement.scrollHeight));await page.locator('.chapter-end').waitFor();await page.waitForFunction(title=>new Promise(resolve=>{const open=indexedDB.open('qmanga-reading');open.onsuccess=()=>{const db=open.result;if(!db.objectStoreNames.contains('positions')){db.close();resolve(false);return;}const request=db.transaction('positions').objectStore('positions').get(['guest',title]);request.onsuccess=()=>{const ready=request.result?.progress.fraction===1;db.close();resolve(ready);};};open.onerror=()=>resolve(false);}),title);assert.equal(await prompt(page).count(),0);assert.equal(suggestions,0);}finally{await guest.close();}
});
