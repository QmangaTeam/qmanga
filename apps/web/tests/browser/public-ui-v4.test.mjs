import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
const f=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(f.app_url).hostname));
test('V4 public UI adapts to locale, theme and viewport while keeping navigation and editors usable',{timeout:240000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL:f.app_url,viewport:{width:390,height:900},reducedMotion:'reduce'});
 try{
  const r=await context.request.post('/api/v1/auth/login',{headers:{Origin:f.app_url},data:{email:f.owner_email,password:f.owner_password}});assert.equal(r.status(),200);const user=(await r.json()).user;
  await context.addInitScript(()=>{const theme=document.cookie.split('; ').find(v=>v.startsWith('qmanga_ui_test_theme='))?.split('=')[1];if(theme)localStorage.setItem('qmanga.theme',theme);});
  const page=await context.newPage();page.setDefaultTimeout(8000);const errors=[];page.on('pageerror',e=>errors.push(e.message));
  const slug=f.title_slug||f.title_id;
  const routes=['/','/catalog','/titles/'+slug+'?tab=about','/titles/'+slug+'?tab=chapters','/titles/'+slug+'?tab=comments','/login','/register','/library','/notifications','/users/'+user.id,'/settings','/settings?section=appearance'];
  for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
   await context.addCookies([{name:'qmanga_locale',value:locale,url:f.app_url},{name:'qmanga_ui_test_theme',value:theme,url:f.app_url}]);
   for(const width of [320,390,834,1280,1440]){
    await page.setViewportSize({width,height:900});
    for(const path of routes){
     await page.goto(path);await page.locator('main').waitFor();await page.evaluate(()=>document.fonts.ready);
     assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),`${locale}/${theme}/${width} ${path}: overflow`);
     assert.equal(await page.locator('html').getAttribute('data-theme'),theme,`${locale}/${theme}/${width} ${path}: actual theme`);
     assert.equal(await page.locator('main').count(),1);
     if(path.includes('?tab=chapters')||path.includes('?tab=comments'))assert.equal(await page.locator('.title-heading .tag-list,.title-heading .synopsis,.about-taxonomy').count(),0);
     if(process.env.QMANGA_SCREENSHOT_DIR&&[390,834,1440].includes(width)){
      await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/${locale}-${theme}-${width}-${routes.indexOf(path)}.png`});
     }
    }
   }
   console.log(`Layout matrix passed: ${locale}/${theme}`);
  }
  await page.setViewportSize({width:390,height:900});await page.goto('/catalog');
  const nav=page.locator('.mobile-navigation');const menu=nav.getByRole('button',{name:'Меню',exact:true});await menu.click();const drawer=page.locator('dialog.navigation-drawer');await drawer.waitFor();
  assert.equal(await drawer.locator('a[href="/settings"],a[href="/admin"]').count(),0,'drawer is site navigation; account actions stay in avatar menu');
  await page.keyboard.press('Escape');await drawer.waitFor({state:'hidden'});assert.equal(await menu.evaluate(e=>e===document.activeElement),true);
  await menu.click();await drawer.locator('a[href="/library"]').click();await page.waitForURL('**/library');assert.equal(await drawer.isVisible(),false);
  await page.goto('/catalog');await page.getByRole('button',{name:'Показать списком',exact:true}).click();assert.equal(await page.locator('.cover-grid.list-view').count(),1);
  await page.getByRole('button',{name:'Фильтры',exact:true}).click();const filters=page.locator('dialog.catalogue-filter-dialog');await filters.waitFor();await page.keyboard.press('Escape');await filters.waitFor({state:'hidden'});
  await page.goto('/titles/'+slug+'?tab=comments');await page.locator('.composer-prompt').click();const composer=page.locator('.comment-composer');
  assert.equal(await composer.locator('.format-actions>button,.emoji-picker>button').count(),2);assert.equal(await composer.locator('.comment-preview').count(),0);
  await composer.getByRole('button',{name:'Спойлер',exact:true}).click();await composer.locator('.spoiler-notice').waitFor();
  await composer.getByRole('button',{name:'Добавить эмодзи',exact:true}).click();await composer.getByRole('button',{name:'Добавить эмодзи: 😊',exact:true}).click();assert.equal(await composer.locator('textarea').inputValue(),'😊');
  await page.goto('/settings');assert.equal(await page.locator('.settings-content').isVisible(),false);await page.locator('.settings-tabs a').first().click();await page.locator('.settings-content').waitFor({state:'visible'});assert.equal(await page.locator('.settings-content').isVisible(),true);
  await page.setViewportSize({width:1440,height:900});await page.goto('/catalog');const previous=await page.locator('html').getAttribute('data-theme');await page.locator('.header-theme button').click();assert.notEqual(await page.locator('html').getAttribute('data-theme'),previous);
  assert.deepEqual(errors,[]);
 }finally{await context.close();await browser.close();}
});
