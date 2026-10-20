import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {dirname,join} from 'node:path';
import {chromium} from 'playwright';

function syntheticPdf(){
 const objects=['<< /Type /Catalog /Pages 2 0 R >>','<< /Type /Pages /Kids [3 0 R 5 0 R] /Count 2 >>',
 '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 300] /Contents 4 0 R >>',
 '<< /Length 26 >>\nstream\n1 0 0 rg 0 0 200 300 re f\nendstream',
 '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 300] /Contents 6 0 R >>',
 '<< /Length 26 >>\nstream\n0 0 1 rg 0 0 200 300 re f\nendstream'];
 let result='%PDF-1.7\n';const offsets=[];
 objects.forEach((value,index)=>{offsets.push(Buffer.byteLength(result));result+=`${index+1} 0 obj\n${value}\nendobj\n`;});
 const start=Buffer.byteLength(result);
 return Buffer.from(result+`xref\n0 7\n0000000000 65535 f \n${offsets.map(offset=>String(offset).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size 7 /Root 1 0 R >>\nstartxref\n${start}\n%%EOF\n`);
}

test('production PDF bundle previews locally and uploads ordered images through the authorized pipeline',{skip:!process.env.QMANGA_FIXTURE_FILE,timeout:90000},async()=>{
 const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
 assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});
 let revision,headers;
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}})).status(),200);
  const token=(await(await context.request.get('/api/v1/auth/csrf')).json()).csrf_token;
  headers={Origin:fixture.app_url,'X-CSRF-Token':token};
  const chapter=await context.request.post('/api/v1/studio/projects/'+fixture.project_id+'/chapters',{headers,data:{label:'PDF-'+Date.now(),volume:'',name:'',order:0}});
  assert.equal(chapter.status(),201);
  const created=await context.request.post('/api/v1/studio/projects/'+fixture.project_id+'/revisions',{headers,data:{chapter_id:(await chapter.json()).id}});
  assert.equal(created.status(),201);revision=(await created.json()).id;
  for(const asset of ['LICENSE','standard_fonts/FoxitSerif.pfb','cmaps/Adobe-Japan1-UCS2.bcmap','wasm/openjpeg.wasm']){
   assert.equal((await context.request.get('/_pdf-assets/'+asset)).status(),200,'self-hosted PDF asset '+asset);
  }
  const page=await context.newPage();let uploads=0;const assetFailures=[];
  page.on('request',request=>{if(request.method()==='PUT'&&request.url().includes('/pages?'))uploads++;});
  page.on('response',response=>{if(response.status()>=400&&(response.url().includes('pdf.worker')||response.url().includes('_pdf-assets')))assetFailures.push(response.status());});
  await page.goto('/studio/projects/'+fixture.project_id+'?revision='+revision);
  const input=page.locator('input[accept*=".pdf"]');await input.waitFor();
  await input.setInputFiles({name:'synthetic.pdf',mimeType:'application/pdf',buffer:syntheticPdf()});
  await page.locator('.import-preview').waitFor();
  assert.deepEqual(await page.locator('.import-preview li').allTextContents(),['pdf-page-001.png','pdf-page-002.png']);
  assert.equal(uploads,0);assert.deepEqual(assetFailures,[]);
  await page.getByRole('button',{name:'Добавить страницы в главу',exact:true}).click();
  await page.waitForFunction(()=>document.querySelectorAll('[id^=studio-page-]').length===2);
  const current=await(await context.request.get('/api/v1/studio/revisions/'+revision)).json();
  assert.deepEqual(current.pages.map(page=>page.original_name),['pdf-page-001.png','pdf-page-002.png']);
  assert.equal(uploads,2);
 }finally{
  try{
   if(revision){const current=await(await context.request.get('/api/v1/studio/revisions/'+revision)).json();await context.request.delete('/api/v1/studio/revisions/'+revision+'?version='+current.version,{headers});}
  }finally{await context.close();await browser.close();}
 }
});

test('cold PDF decoder download can be cancelled before chunks finish without changing the draft',{skip:!process.env.QMANGA_FIXTURE_FILE,timeout:60000},async()=>{
 const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:360,height:850}});let revision,headers,release=()=>{};
 try{
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);assert.equal((await context.request.post('/api/v1/auth/login',{headers:{Origin:fixture.app_url},data:{email:fixture.owner_email,password:fixture.owner_password}})).status(),200);
  headers={Origin:fixture.app_url,'X-CSRF-Token':(await(await context.request.get('/api/v1/auth/csrf')).json()).csrf_token};
  const chapter=await context.request.post('/api/v1/studio/projects/'+fixture.project_id+'/chapters',{headers,data:{label:'PDF-cancel-'+Date.now(),volume:'',name:'',order:0}});assert.equal(chapter.status(),201);
  const created=await context.request.post('/api/v1/studio/projects/'+fixture.project_id+'/revisions',{headers,data:{chapter_id:(await chapter.json()).id}});assert.equal(created.status(),201);revision=(await created.json()).id;
  const image=await readFile(join(dirname(process.env.QMANGA_FIXTURE_FILE),'pages','page-01.png'));
  assert.equal((await context.request.put('/api/v1/studio/revisions/'+revision+'/pages?key=cold-cancel-existing&position=1&name=existing.png',{headers:{...headers,'Content-Type':'application/octet-stream'},data:image})).status(),201);
  const before=await(await context.request.get('/api/v1/studio/revisions/'+revision)).json();assert.equal(before.pages.length,1);
  const page=await context.newPage();await page.goto('/studio/projects/'+fixture.project_id+'?revision='+revision);const input=page.locator('input[accept*=".pdf"]');await input.waitFor();
  let hit;const requested=new Promise(resolve=>hit=resolve),held=new Promise(resolve=>release=resolve);let uploads=0;page.on('request',r=>{if(r.method()==='PUT'&&r.url().includes('/pages?'))uploads++;});
  await page.route('**/_app/immutable/chunks/*.js',async route=>{hit();await held;await route.continue().catch(()=>{});});
  await input.setInputFiles({name:'cold.pdf',mimeType:'application/pdf',buffer:syntheticPdf()});await requested;
  await page.getByRole('button',{name:'Остановить',exact:true}).click();await page.waitForFunction(()=>!document.querySelector('input[accept*=".pdf"]').disabled,{},{timeout:2000});
  assert.equal(await page.getByText('Подготовка страниц',{exact:false}).count(),0);assert.equal(await page.locator('.import-preview').count(),0);assert.equal(uploads,0);
  const current=await(await context.request.get('/api/v1/studio/revisions/'+revision)).json();assert.deepEqual(current.pages.map(p=>p.id),before.pages.map(p=>p.id));
  release();await page.waitForTimeout(300);assert.equal(await page.locator('.import-preview').count(),0);assert.equal(uploads,0);
 }finally{release();try{if(revision){const current=await(await context.request.get('/api/v1/studio/revisions/'+revision)).json();await context.request.delete('/api/v1/studio/revisions/'+revision+'?version='+current.version,{headers});}}finally{await context.close();await browser.close();}}
});
