import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile,mkdir} from 'node:fs/promises';
import {chromium} from 'playwright';
if(!process.env.QMANGA_FIXTURE_FILE)throw new Error('QMANGA_FIXTURE_FILE required');
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
if(!['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname))throw new Error('Local fixtures only');
test('status updates the title button immediately and ratings save from ten accessible choices',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{
  const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
  try{
   await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
   const login=await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}});assert.equal(login.status(),200);
   const page=await context.newPage();page.setDefaultTimeout(7000);await page.goto('/titles/'+(fixture.title_slug||fixture.title_id));
   await page.locator('.library-trigger').click();const library=page.locator('.action-dialog').filter({has:page.locator('.selection')});
   await library.getByRole('button',{name:'Брошено',exact:true}).click();
   await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Брошено'));
   const boxes=await library.locator('.status-choices button').evaluateAll(elements=>elements.map(e=>e.getBoundingClientRect()));
   assert.ok(boxes.every(b=>Math.abs(b.x-boxes[0].x)<1),'statuses form one column');
   await library.getByRole('button',{name:'Закрыть',exact:true}).click();
   await page.reload();await page.waitForFunction(()=>document.querySelector('.library-trigger')?.textContent.includes('Брошено'));
   await page.locator('.rating-trigger').click();const dialog=page.locator('.rating-dialog');
   const choices=dialog.locator('.rating-choices button');assert.equal(await choices.count(),10);
   assert.equal(await dialog.getByRole('combobox').count(),0);assert.equal(await dialog.getByRole('button',{name:'Сохранить',exact:true}).count(),0);
   await choices.nth(6).click();await dialog.waitFor({state:'hidden'});
   assert.equal(await page.locator('.rating-trigger .rate-label').textContent(),'Ваша: 7');
   await page.locator('.rating-trigger').click();assert.equal(await choices.nth(6).getAttribute('aria-pressed'),'true');
   let hoverWrites=0;const observe=request=>{if(request.method()==='PUT'&&request.url().endsWith('/rating'))hoverWrites++;};page.on('request',observe);
   await choices.nth(1).hover();assert.equal(await dialog.locator('.rating-number').textContent(),'2');assert.equal(await dialog.locator('.rating-phrase').textContent(),'Слабое впечатление');assert.equal(await dialog.locator('svg.filled').count(),2);
   await dialog.locator('h2').hover();assert.equal(await dialog.locator('.rating-number').textContent(),'7');assert.equal(hoverWrites,0);page.off('request',observe);

   await page.route('**/api/v1/community/titles/*/rating',async route=>route.request().method()==='PUT'?route.fulfill({status:500,json:{error:{code:'internal'}}}):route.continue());
   await choices.nth(7).click();await dialog.getByRole('alert').waitFor();assert.equal(await dialog.isVisible(),true);assert.equal(await choices.nth(6).getAttribute('aria-pressed'),'true');
   await page.unroute('**/api/v1/community/titles/*/rating');
   await dialog.getByRole('button',{name:'Закрыть',exact:true}).click();await page.locator('.rating-trigger').click();await page.waitForFunction(()=>document.querySelector('.rating-choices button')?.disabled===false);
   for(const theme of ['dark','light'])for(const width of [320,360,1280]){
    await page.evaluate(value=>document.documentElement.dataset.theme=value,theme);await page.setViewportSize({width,height:850});
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
    const size=await choices.first().boundingBox();assert.ok(size.height>=44);const row=await dialog.locator('.rating-choices').evaluate(el=>({width:el.clientWidth,scroll:el.scrollWidth}));assert.ok(row.scroll<=row.width,'all ten ratings fit without horizontal scrolling');
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/rating-${theme}-${width}.png`});}
   }

   const titleURL='/titles/'+(fixture.title_slug||fixture.title_id);
   const beforeCount=(await (await context.request.get('/api/v1/titles/'+(fixture.title_slug||fixture.title_id))).json()).comment_count;
   await page.goto(titleURL+'?tab=comments');const draft='Synthetic counter test '+Date.now();
   await page.locator('.composer-prompt').click();
   const composer=page.locator('.discussion .comment-composer');await composer.getByRole('textbox',{name:'Написать комментарий',exact:true}).fill(draft);
   await composer.getByRole('button',{name:'Отправить',exact:true}).click();
   await page.waitForFunction(expected=>Number(document.querySelector('.title-tabs a[href*="tab=comments"] .tab-count')?.textContent)===expected,beforeCount+1);
   const posted=page.locator('.discussion>article').filter({hasText:draft});await posted.getByRole('button',{name:'Удалить',exact:true}).click();
   await page.waitForFunction(expected=>Number(document.querySelector('.title-tabs a[href*="tab=comments"] .tab-count')?.textContent)===expected,beforeCount);
   const comments=Array.from({length:20},(_,index)=>({id:'fixture-'+index,user_id:'other',display_name:'Synthetic reader',parent_id:'',root_id:'',body:('Synthetic discussion line '+index+'\n').repeat(6),spoiler:false,hidden:false,created_at:new Date(Date.UTC(2026,0,1,0,index)).toISOString(),replies:0}));
   await page.route('**/api/v1/community/titles/*?*',route=>route.fulfill({json:{comments,next:'',score:0,ratings:0}}));
   await page.setViewportSize({width:360,height:850});await page.goto('/titles/'+(fixture.title_slug||fixture.title_id)+'?tab=comments');
   const root=page.locator('#comment-fixture-0');await root.scrollIntoViewIfNeeded();const before=await page.evaluate(()=>scrollY);const beforeTop=(await root.boundingBox()).y;
   await root.getByRole('button',{name:'Ответить',exact:true}).click();
   const reply=root.getByRole('textbox',{name:'Написать ответ',exact:true});await reply.waitFor();
   const after=await page.evaluate(()=>scrollY);const afterTop=(await root.boundingBox()).y;assert.ok(Math.abs(afterTop-beforeTop)<48,JSON.stringify({before,after,beforeTop,afterTop}));
   await reply.fill('Draft stays beside this comment');
   assert.equal(await reply.evaluate(el=>el===document.activeElement),true);
   await root.getByRole('button',{name:'Отмена',exact:true}).click();assert.equal(await root.locator('textarea').count(),0);
   comments.push({...comments[0],id:'fixture-reply',parent_id:'fixture-0',root_id:'fixture-0',body:'Synthetic reply'});
   await page.reload();await page.locator('#comment-fixture-0 .thread-toggle').click();const nested=page.locator('#comment-fixture-0 .reply');await nested.scrollIntoViewIfNeeded();
   await nested.getByRole('button',{name:'Ответить',exact:true}).click();await nested.getByRole('textbox',{name:'Написать ответ',exact:true}).fill('Reply to reply remains here');
   assert.equal(await page.locator('.discussion [id^="root-editor-"] form').count(),0);
   await nested.getByRole('button',{name:'Отмена',exact:true}).click();

  }finally{await context.close();}
 }finally{await browser.close();}
});
