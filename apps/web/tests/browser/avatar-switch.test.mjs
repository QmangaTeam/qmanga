import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Dedicated local fixture required');
const actor=randomUUID(),teams=[randomUUID(),randomUUID()],email=`avatar-switch-${actor}@test.local`;
let browser;
function sql(input){const result=spawnSync('docker',['compose','--env-file',fixture.env_file,'-p',fixture.project_name,'exec','-T','postgres','psql','-U','qmanga','-d','qmanga','-v','ON_ERROR_STOP=1'],{input,encoding:'utf8',timeout:30000});assert.equal(result.status,0,result.stderr);}
before(async()=>{assert.match(fixture.owner_email,/^[a-zA-Z0-9@._+-]+$/);sql(`BEGIN;INSERT INTO users(id,email,password_hash,is_verified,display_name) SELECT '${actor}','${email}',password_hash,true,'Avatar switch' FROM users WHERE email='${fixture.owner_email}';`+teams.map((id,index)=>`INSERT INTO teams(id,slug,name,is_approved)VALUES('${id}','avatar-switch-${id}','Avatar ${index}',true);INSERT INTO team_memberships(id,team_id,user_id,role)VALUES('${randomUUID()}','${id}','${actor}','owner');`).join('')+'COMMIT;');browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});});
after(async()=>{try{
 // Keep storage deletion under the ordinary durable queue, including fixture cleanup.
 sql(`BEGIN;INSERT INTO avatar_garbage(object_key,ready_at) SELECT object_key,now() FROM avatars WHERE owner_kind='team' AND owner_id IN('${teams[0]}','${teams[1]}') ON CONFLICT DO NOTHING;DELETE FROM avatars WHERE owner_kind='team' AND owner_id IN('${teams[0]}','${teams[1]}');DELETE FROM team_memberships WHERE user_id='${actor}';DELETE FROM teams WHERE id IN('${teams[0]}','${teams[1]}');DELETE FROM users WHERE id='${actor}';COMMIT;`);
 }finally{await browser?.close();}});
test('team avatar fallback and pending upload remain scoped to the selected team',{timeout:30000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});let releaseUpload;
 try{await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email,password:fixture.owner_password}})).status(),200);
 const csrf=(await(await context.request.get('/api/v1/auth/csrf')).json()).csrf_token;const headers={Origin:baseURL,'X-Qmanga-Account':actor,'X-CSRF-Token':csrf,'Content-Type':'image/png'};
 const page=await context.newPage();page.setDefaultTimeout(8000);await page.goto('/studio/teams');
 const png=Buffer.from(await page.evaluate(()=>{const canvas=document.createElement('canvas');canvas.width=320;canvas.height=180;const ctx=canvas.getContext('2d');ctx.fillStyle='#ff873d';ctx.fillRect(0,0,320,180);return canvas.toDataURL('image/png').split(',')[1];}),'base64');
 assert.equal((await context.request.post('/api/v1/avatars/team/'+teams[1],{headers,data:png})).status(),200);
 const selector=page.locator('.team-select select');await selector.selectOption(teams[0]);await page.locator('.avatar-editor .avatar img').waitFor({state:'detached'});
 await selector.selectOption(teams[1]);await page.waitForFunction(id=>{const img=document.querySelector('.avatar-editor .avatar img');return img?.src.includes(id)&&img.naturalWidth===256;},teams[1]);
 assert.equal(await page.locator('.avatar-editor img').count(),1);
 await selector.selectOption(teams[0]);await page.locator('.avatar-editor img').waitFor({state:'detached'});
 let observed;const requestObserved=new Promise(resolve=>{observed=resolve;});const delayed=new Promise(resolve=>{releaseUpload=resolve;});
 await page.route('**/api/v1/avatars/team/'+teams[0],async route=>{observed();await delayed;await route.continue();});
 await page.locator('.avatar-editor input').setInputFiles({name:'team.png',mimeType:'image/png',buffer:png});await page.getByRole('button',{name:'Сохранить',exact:true}).click();await requestObserved;await selector.selectOption(teams[1]);
 const completed=page.waitForResponse(response=>response.url().endsWith('/avatars/team/'+teams[0])&&response.request().method()==='POST');releaseUpload();assert.equal((await completed).status(),200);
 await page.waitForFunction(id=>{const img=document.querySelector('.avatar-editor img');return img?.src.includes(id)&&img.naturalWidth===256;},teams[1]);assert.equal(await page.getByRole('status').filter({hasText:'Аватар сохранён.'}).count(),0,'old upload must not announce success for a different team');
 assert.equal(await page.locator('.avatar-editor input').isDisabled(),false);
 }finally{releaseUpload?.();await context.close();}
});
