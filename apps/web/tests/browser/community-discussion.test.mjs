import {before,after,test} from 'node:test';
import assert from 'node:assert/strict';
import {spawn} from 'node:child_process';
import {createServer} from 'node:net';
import {mkdir} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import {chromium} from 'playwright';

// Isolated presentation fixture: its in-memory gateway never calls the real API.
let browser,server,baseURL;
before(async()=>{
 const socket=createServer();await new Promise(resolve=>socket.listen(0,'127.0.0.1',resolve));const port=socket.address().port;await new Promise(resolve=>socket.close(resolve));
 baseURL='http://127.0.0.1:'+port;
 const web=fileURLToPath(new URL('../../',import.meta.url));
 server=spawn(process.execPath,[web+'node_modules/vite/bin/vite.js','--config',web+'tests/community-preview/community-preview.config.ts','--port',String(port)],{cwd:web,stdio:['ignore','pipe','pipe']});
 let log='';server.stdout.on('data',chunk=>log+=chunk);server.stderr.on('data',chunk=>log+=chunk);
 let ready=false;for(let attempt=0;attempt<150;attempt++){try{ready=(await fetch(baseURL)).ok;if(ready)break;}catch{}await new Promise(resolve=>setTimeout(resolve,100));}
 assert.ok(ready,'Community preview starts: '+log);
 browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
});
after(async()=>{await browser?.close();server?.kill();});
const calls=async page=>JSON.parse(await page.getByTestId('calls').textContent());
async function open(page,query='locale=ru'){
 await page.route('**/media/avatars/**',route=>route.fulfill({status:404,body:''}));
 page.setDefaultTimeout(7000);await page.goto(baseURL+'/?'+query);await page.locator('#comment-root').waitFor();
}

test('comment hierarchy and contextual composer fit both themes, languages and widths',{timeout:60000},async()=>{
 const context=await browser.newContext({reducedMotion:'reduce'});const page=await context.newPage();const errors=[];page.on('pageerror',error=>errors.push(error.message));
 try{
  for(const locale of ['kk','ru'])for(const theme of ['dark','light']){
   await open(page,`locale=${locale}&theme=${theme}`);
   for(const width of [320,360,1280]){
    await page.setViewportSize({width,height:900});
    const root=page.locator('#comment-root');
    await root.locator('.comment-actions button').first().click();
    await page.locator('.comment-composer textarea').fill('ӨтеҰзақСөз'.repeat(80));
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),`${locale}/${theme}/${width} page fits: `+JSON.stringify(await page.locator('body *').evaluateAll(elements=>elements.filter(element=>element.getBoundingClientRect().right>innerWidth+1).map(element=>({tag:element.tagName,class:element.className,right:element.getBoundingClientRect().right})).slice(0,15))));
    assert.equal(await page.locator('.comment-composer textarea').evaluate(element=>element===document.activeElement),true);
    assert.equal(await page.locator('.comment-header time').count(),3,'replies have dates as well as the root');
    assert.ok(await page.locator('.reply-address').isVisible(),'nested reply names its parent');
    assert.equal(await page.locator('.avatar img').count(),0,'missing avatars settle to initials');
    for(const box of await page.locator('.comment-actions button,.comment-composer button,.comment-composer summary,.more-replies').evaluateAll(elements=>elements.map(element=>({width:element.getBoundingClientRect().width,height:element.getBoundingClientRect().height}))))assert.ok(box.width>=44&&box.height>=44,'interactive targets remain at least 44px');
    if(process.env.QMANGA_SCREENSHOT_DIR){await mkdir(process.env.QMANGA_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:`${process.env.QMANGA_SCREENSHOT_DIR}/discussion-${locale}-${theme}-${width}.png`,fullPage:true});}
    await page.locator('.close-composer').click();
    assert.equal(await root.locator('.comment-actions button').first().evaluate(element=>element===document.activeElement),true,'close returns focus to reply action');
   }
  }
  assert.deepEqual(errors,[]);
 }finally{await context.close();}
});

test('reply and edit retain formatting, spoilers, preview, cancellation and exact mutation scope',{timeout:30000},async()=>{
 const page=await browser.newPage({viewport:{width:320,height:850}});
 try{
  await open(page);
  await page.locator('#comment-root .thread-toggle').click();
  const nested=page.locator('#comment-nested');await nested.getByRole('button',{name:'Ответить',exact:true}).click();
  const composer=page.locator('.comment-composer'),editor=composer.getByRole('textbox');
  assert.match(await composer.locator('.composer-context').textContent(),/Ответ для Оқырман/);
  await editor.fill('Ответ с форматированием');await editor.evaluate(element=>element.setSelectionRange(0,5));
  await composer.getByRole('button',{name:'Жирный',exact:true}).click();assert.equal(await editor.inputValue(),'**Ответ** с форматированием');
  await composer.getByRole('button',{name:'Спойлер',exact:true}).click();assert.equal(await composer.getByRole('button',{name:'Спойлер',exact:true}).getAttribute('aria-pressed'),'true');
  await composer.getByText('Предпросмотр',{exact:true}).click();assert.equal(await composer.locator('.comment-preview strong').textContent(),'Ответ');
  await composer.getByRole('button',{name:'Отправить',exact:true}).click();await page.locator('#comment-new').waitFor();
  assert.deepEqual((await calls(page)).find(call=>call.kind==='comment'),{kind:'comment',account:'own',title:'synthetic-title',release:'',parent:'nested',body:'**Ответ** с форматированием',spoiler:true});
  const reply=page.locator('#comment-reply');await reply.getByRole('button',{name:'Редактировать',exact:true}).click();
  assert.equal(await editor.inputValue(),'Менің жауабым');await editor.fill('Отменённый черновик');
  await composer.getByRole('button',{name:'Отмена',exact:true}).click();assert.equal((await calls(page)).filter(call=>call.kind==='edit').length,0);
  assert.equal(await reply.getByRole('button',{name:'Редактировать',exact:true}).evaluate(element=>element===document.activeElement),true);
  await reply.getByRole('button',{name:'Редактировать',exact:true}).click();await editor.fill('Изменённый ответ');
  await composer.getByRole('button',{name:'Отправить',exact:true}).click();await page.getByText('Изменённый ответ',{exact:true}).waitFor();
  assert.equal((await calls(page)).find(call=>call.kind==='edit').id,'reply');
  await reply.getByRole('button',{name:'Удалить',exact:true}).click();await reply.waitFor({state:'hidden'});
  assert.equal((await calls(page)).find(call=>call.kind==='remove').id,'reply');
 }finally{await page.close();}
});

test('reports, reply paging, comment paging and focused links keep their targets',{timeout:30000},async()=>{
 const page=await browser.newPage();
 try{
  await open(page,'locale=ru&comment=nested');
  assert.equal(await page.locator('#comment-nested').evaluate(element=>element===document.activeElement),true);
  await page.getByRole('button',{name:'Показать остальные ответы',exact:true}).click();await page.locator('#comment-remaining').waitFor();
  assert.equal((await calls(page)).find(call=>call.kind==='replies').id,'root');
  await page.getByRole('button',{name:'Следующие комментарии',exact:true}).click();await page.locator('#comment-older').waitFor();
  assert.equal((await calls(page)).filter(call=>call.kind==='discussion').at(-1).after,'older-page');
  const reply=page.locator('#comment-reply');await reply.getByRole('button',{name:'Пожаловаться',exact:true}).click();
  await reply.getByText('Сообщить о проблеме',{exact:true}).click();await reply.getByRole('textbox').fill('Подробная причина жалобы');
  await reply.getByRole('button',{name:'Отправить жалобу',exact:true}).click();
  await page.waitForFunction(()=>JSON.parse(document.querySelector('[data-testid="calls"]').textContent).some(call=>call.kind==='report'));
  const report=(await calls(page)).find(call=>call.kind==='report');assert.equal(report.targetType,'comment');assert.equal(report.targetId,'reply');
 }finally{await page.close();}
});

test('guest, empty discussion and failed-load retry remain available',{timeout:30000},async()=>{
 const page=await browser.newPage();
 try{
  await open(page,'locale=ru&guest=1');assert.equal(await page.locator('.comment-composer').count(),0);assert.equal(await page.getByRole('link',{name:'войди',exact:true}).count(),1);
  await page.goto(baseURL+'/?locale=ru&empty=1');await page.getByText('Оставь первый комментарий.',{exact:true}).waitFor();assert.equal(await page.locator('.comment-composer').count(),0);assert.equal(await page.locator('.composer-prompt').count(),1);
  await page.goto(baseURL+'/?locale=ru&error=1');await page.getByRole('button',{name:'Повторить',exact:true}).click();await page.locator('#comment-root').waitFor();
  assert.equal((await calls(page)).filter(call=>call.kind==='discussion').length,2);
 }finally{await page.close();}
});


test('a failed reply keeps its draft, recipient and spoiler toggle for retry',{timeout:30000},async()=>{
 const page=await browser.newPage({viewport:{width:360,height:850}});
 try{
  await open(page,'locale=ru&submit-error=1');
  await page.locator('#comment-root .thread-toggle').click();
  const nested=page.locator('#comment-nested');
  await nested.getByText('Показать спойлер',{exact:true}).click();assert.equal(await nested.getByText('Жасырын жауап',{exact:true}).isVisible(),true);
  await nested.getByRole('button',{name:'Ответить',exact:true}).click();
  const composer=page.locator('.comment-composer'),editor=composer.getByRole('textbox');
  await editor.fill('Черновик после ошибки');await composer.getByRole('button',{name:'Спойлер',exact:true}).click();
  await composer.getByRole('button',{name:'Отправить',exact:true}).click();await page.locator('.form-message').waitFor();
  assert.equal(await editor.inputValue(),'Черновик после ошибки');assert.match(await composer.locator('.composer-context').textContent(),/Ответ для Оқырман/);
  assert.equal(await composer.getByRole('button',{name:'Спойлер',exact:true}).getAttribute('aria-pressed'),'true');
  await composer.getByRole('button',{name:'Отправить',exact:true}).click();await page.locator('#comment-new').waitFor();
  assert.equal((await calls(page)).find(call=>call.kind==='comment').parent,'nested');
 }finally{await page.close();}
});

test('root composer and loaded replies disclose without losing drafts or refetching',{timeout:30000},async()=>{
 const page=await browser.newPage({viewport:{width:360,height:844},reducedMotion:'reduce'});
 try{
  await open(page);
  const prompt=page.locator('.composer-prompt'),root=page.locator('#comment-root'),toggle=root.locator('.thread-toggle');
  assert.equal(await page.locator('.comment-composer').count(),0);
  assert.equal(await toggle.getAttribute('aria-expanded'),'false');
  assert.equal(await page.locator('#comment-reply').isVisible(),false);
  await prompt.focus();await page.keyboard.press('Enter');
  const editor=page.locator('.comment-composer textarea');await editor.fill('Сохранённый черновик');
  await page.getByRole('button',{name:'Спойлер',exact:true}).click();
  await page.locator('.close-composer').click();assert.equal(await prompt.evaluate(e=>e===document.activeElement),true);
  await prompt.click();assert.equal(await editor.inputValue(),'Сохранённый черновик');assert.equal(await page.getByRole('button',{name:'Спойлер',exact:true}).getAttribute('aria-pressed'),'true');
  await root.locator('.comment-actions button').first().click();await editor.fill('Ответ в ветке');await page.locator('.close-composer').click();
  assert.equal(await editor.inputValue(),'Сохранённый черновик','root draft restored after inline cancellation');
  await page.locator('.close-composer').click();
  await toggle.click();assert.equal(await toggle.getAttribute('aria-expanded'),'false');
  await toggle.focus();await page.keyboard.press('Space');await page.locator('#comment-reply').waitFor({state:'visible'});
  await toggle.click();await toggle.click();assert.equal((await calls(page)).filter(c=>c.kind==='replies').length,0,'loaded replies do not refetch');
  await page.getByRole('button',{name:'Показать остальные ответы',exact:true}).click();await page.locator('#comment-remaining').waitFor();
  await toggle.click();await toggle.click();assert.equal((await calls(page)).filter(c=>c.kind==='replies').length,1);
 }finally{await page.close();}
});

test('failed root send preserves editor and successful retry clears the draft',{timeout:30000},async()=>{
 const page=await browser.newPage();
 try{await open(page,'locale=ru&submit-error=1');await page.locator('.composer-prompt').click();
  const editor=page.locator('.comment-composer textarea');await editor.fill('Корневой черновик после ошибки');await page.getByRole('button',{name:'Спойлер',exact:true}).click();
  await page.getByRole('button',{name:'Отправить',exact:true}).click();await page.locator('.form-message').waitFor();
  assert.equal(await editor.inputValue(),'Корневой черновик после ошибки');assert.equal(await page.getByRole('button',{name:'Спойлер',exact:true}).getAttribute('aria-pressed'),'true');
  await page.getByRole('button',{name:'Отправить',exact:true}).click();await page.locator('#comment-new').waitFor();await page.locator('.composer-prompt').waitFor();
  assert.equal(await page.locator('.comment-composer').count(),0);assert.equal(await page.locator('.composer-prompt').evaluate(e=>e===document.activeElement),true);
  await page.locator('.composer-prompt').click();assert.equal(await editor.inputValue(),'');
 }finally{await page.close();}
});
