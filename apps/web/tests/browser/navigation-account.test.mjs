import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {randomUUID} from 'node:crypto';
import {spawnSync} from 'node:child_process';
import {join} from 'node:path';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8')),baseURL=fixture.app_url;
if(!['localhost','127.0.0.1'].includes(new URL(baseURL).hostname)||!/^qmanga-(ci$|stage[0-9]+-)/.test(fixture.project_name||''))throw Error('Dedicated local integration fixture required');
let browser,adminState;const userId=randomUUID(),email=`nav-${userId}@test.local`,prefix=randomUUID().slice(0,30);
function sql(input){const result=spawnSync('docker',['compose','--env-file',fixture.env_file,'-p',fixture.project_name,'exec','-T','postgres','psql','-U','qmanga','-d','qmanga','-v','ON_ERROR_STOP=1'],{input,encoding:'utf8',timeout:30000});if(result.status!==0)throw Error(result.stderr);}
async function login(context,address=fixture.admin_email,password=fixture.admin_password){const result=await context.request.post('/api/v1/auth/login',{headers:{Origin:baseURL},data:{email:address,password}});assert.equal(result.status(),200);return (await result.json()).user;}
before(async()=>{browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 // Copy the disposable admin's password hash to a new reader; no existing account/data is modified.
 assert.match(fixture.admin_email,/^[a-zA-Z0-9@._+-]+$/);
 sql(`INSERT INTO users(id,email,password_hash,display_name,is_verified) SELECT '${userId}','${email}',password_hash,'Navigation Reader',true FROM users WHERE email='${fixture.admin_email}'; INSERT INTO notifications(id,user_id,type,title_kk,body_kk) SELECT '${prefix}'||n,'${userId}','comment_reply','Сынақ','Navigation fixture' FROM generate_series(1,105)n;`);
 const admin=await browser.newContext({baseURL});try{await login(admin);adminState=await admin.storageState();}finally{await admin.close();}
});
after(async()=>{try{sql(`DELETE FROM notifications WHERE user_id='${userId}';DELETE FROM users WHERE id='${userId}';`);}finally{await browser?.close();}});
for(const locale of ['kk','ru'])for(const theme of ['light','dark'])test(`account and separate admin navigation remain usable: ${locale}/${theme}`,{timeout:45000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:1280,height:800},reducedMotion:'reduce',storageState:adminState});
 try{
 await context.addCookies([{name:'qmanga_locale',value:locale,url:baseURL}]);await context.addInitScript(value=>localStorage.setItem('qmanga.theme',value),theme);
 const page=await context.newPage();page.setDefaultTimeout(8000);await page.goto('/menu');
 const management=page.locator('.menu-page').getByRole('link',{name:locale==='ru'?'Управление':'Басқару',exact:true});assert.equal(await management.count(),1);
 const trigger=page.getByRole('button',{name:locale==='ru'?'Меню аккаунта':'Аккаунт мәзірі',exact:true});await trigger.click();const menu=page.locator('dialog.account-menu');await menu.waitFor();assert.equal(await menu.getByRole('link',{name:locale==='ru'?'Управление':'Басқару',exact:true}).count(),1);await page.keyboard.press('Escape');await menu.waitFor({state:'hidden'});assert.equal(await trigger.evaluate(el=>el===document.activeElement),true);
 await management.click();await page.waitForURL('**/admin');assert.equal(await page.locator('.admin-sidebar').isVisible(),true);assert.equal(await page.locator('.primary-navigation').count(),0);assert.equal(await page.getByRole('link',{name:locale==='ru'?'Журнал аудита':'Аудит журналы',exact:true}).count(),1);
 for(const width of [320,360]){await page.setViewportSize({width,height:750});assert.equal(await page.locator('.admin-sidebar').isVisible(),false);await page.getByRole('button',{name:locale==='ru'?'Разделы':'Бөлімдер',exact:true}).click();const nav=page.getByRole('dialog',{name:locale==='ru'?'Управление':'Басқару',exact:true});await nav.waitFor();assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true);if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:join(process.env.QMANGA_SCREENSHOT_DIR,`admin-nav-${locale}-${theme}-${width}.png`)});}await page.keyboard.press('Escape');await nav.waitFor({state:'hidden'});}
 await page.goto('/');await page.setViewportSize({width:1280,height:800});assert.equal(await page.locator('link[rel="icon"]').count(),1);assert.match(await page.locator('link[rel="icon"]').getAttribute('href'),/\/favicon\.png$/);assert.equal(await page.locator('link[rel="apple-touch-icon"]').count(),1);assert.match(await page.locator('link[rel="apple-touch-icon"]').getAttribute('href'),/\/apple-touch-icon\.png$/);assert.equal(Number.parseFloat(await page.locator('header a.wordmark img').first().evaluate(el=>getComputedStyle(el).height)),40);const footer=page.locator('footer.site-footer');assert.equal(await footer.locator('section').count(),4);for(const href of ['/settings','/reports','/help','/help#publishing','/legal/terms','/legal/copyright'])assert.equal(await footer.locator(`a[href="${href}"]`).count(),1);
 for(const policy of ['terms','copyright','privacy']){const result=await page.goto('/legal/'+policy);assert.equal(result.status(),200);assert.equal(await page.locator('.draft').count(),1);assert.equal(await page.locator('a[href^="mailto:"]').count(),0);assert.equal(await page.locator('.policy').getByText('qmanga-support@maqsatto.dev',{exact:true}).count(),1);}
 }finally{await context.close();}
});
test('unread count is exact, refreshes after read and focus, and guests/readers cannot enter admin',{timeout:30000},async()=>{
 const context=await browser.newContext({baseURL,viewport:{width:360,height:800}});try{
 await context.addCookies([{name:'qmanga_locale',value:'ru',url:baseURL}]);await login(context,email);const page=await context.newPage();await page.goto('/notifications');
 const bell=page.getByRole('navigation',{name:'Мобильная навигация'}).getByRole('link',{name:'Уведомления: 105',exact:true});await bell.waitFor();assert.equal(await bell.getByText('99+').count(),1);
 await page.getByRole('button',{name:'Отметить прочитанным'}).first().click();
 await page.getByRole('navigation',{name:'Мобильная навигация'}).getByRole('link',{name:'Уведомления: 104',exact:true}).waitFor();
 sql(`DELETE FROM notifications WHERE user_id='${userId}' AND id NOT IN (SELECT id FROM notifications WHERE user_id='${userId}' AND NOT is_read ORDER BY id LIMIT 4);`);
 await page.evaluate(()=>window.dispatchEvent(new Event('focus')));await page.getByRole('navigation',{name:'Мобильная навигация'}).getByRole('link',{name:'Уведомления: 4',exact:true}).waitFor();
 assert.equal((await page.goto('/admin')).status(),403);await page.goto('/menu');assert.equal(await page.locator('.menu-page a[href="/admin"]').count(),0);
 }finally{await context.close();}
 const guest=await browser.newContext({baseURL});try{const page=await guest.newPage();await page.goto('/');assert.equal(await page.locator('.unread-badge').count(),0);await page.goto('/admin');assert.match(new URL(page.url()).pathname,/login/);assert.equal((await guest.request.get('/api/v1/notifications/unread-count')).status(),401);}finally{await guest.close();}
});
