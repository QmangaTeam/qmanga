import { before, after, test } from 'node:test';
import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';
import { createServer } from 'vite';
import { pdfAssets } from '../../scripts/pdf-assets.mjs';

let server, browser, base;
before(async () => {
 server = await createServer({ configFile:false, plugins:[pdfAssets()], appType:'custom', root: fileURLToPath(new URL('../../', import.meta.url)), resolve: { alias: { '$lib': fileURLToPath(new URL('../../src/lib', import.meta.url)) } }, server: { host: '127.0.0.1', port: 0 } });
 server.middlewares.use('/__pdf_import_test', (_req,res) => { res.setHeader('Content-Type','text/html');res.end('<!doctype html><title>PDF codec test</title>'); });
 await server.listen(); base = `http://127.0.0.1:${server.httpServer.address().port}`;
 browser = await chromium.launch({ headless:true, ...(process.env.QMANGA_BROWSER_EXECUTABLE ? {executablePath:process.env.QMANGA_BROWSER_EXECUTABLE} : {}) });
});
after(async () => { await browser?.close(); await server?.close(); });

// Original synthetic vector PDF: two colored pages, with valid offsets and no fixture assets.
function pdf() {
 const objects = [ '<< /Type /Catalog /Pages 2 0 R >>', '<< /Type /Pages /Kids [3 0 R 5 0 R] /Count 2 >>',
  '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 300] /Contents 4 0 R >>',
  '<< /Length 30 >>\nstream\n1 0 0 rg 0 0 200 300 re f\nendstream',
  '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 300] /Contents 6 0 R >>',
  '<< /Length 30 >>\nstream\n0 0 1 rg 0 0 200 300 re f\nendstream' ];
 let result = '%PDF-1.7\n'; const offsets=[0];
 objects.forEach((value,index)=>{offsets.push(Buffer.byteLength(result));result+=`${index+1} 0 obj\n${value}\nendobj\n`;});
 const xref=Buffer.byteLength(result); result+=`xref\n0 7\n0000000000 65535 f \n${offsets.slice(1).map(offset=>String(offset).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size 7 /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`;
 return result;
}

// Standard PDF encryption revision 2, generated locally with a nonempty user password.
function encryptedPdf() {
 const pad=Buffer.from('28bf4e5e4e758a4164004e56fffa01082e2e00b6d0683e802f0ca9fe6453697a','hex');
 const padded=text=>Buffer.concat([Buffer.from(text),pad]).subarray(0,32);
 const md5=data=>createHash('md5').update(data).digest();
 const rc4=(key,data)=>{const state=Uint8Array.from({length:256},(_,i)=>i);let j=0;for(let i=0;i<256;i++){j=(j+state[i]+key[i%key.length])&255;[state[i],state[j]]=[state[j],state[i]];}let i=0;j=0;return Buffer.from(data.map(byte=>{i=(i+1)&255;j=(j+state[i])&255;[state[i],state[j]]=[state[j],state[i]];return byte^state[(state[i]+state[j])&255];}));};
 const owner=rc4(md5(padded('owner')).subarray(0,5),padded('secret'));
 const id=Buffer.from('0123456789abcdef0123456789abcdef','hex');const permission=Buffer.alloc(4);permission.writeInt32LE(-4);
 const key=md5(Buffer.concat([padded('secret'),owner,permission,id])).subarray(0,5);const user=rc4(key,pad);
 const objects=['<< /Type /Catalog /Pages 2 0 R >>','<< /Type /Pages /Kids [3 0 R] /Count 1 >>','<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 300] >>',`<< /Filter /Standard /V 1 /R 2 /Length 40 /O <${owner.toString('hex')}> /U <${user.toString('hex')}> /P -4 >>`];
 let result='%PDF-1.7\n';const offsets=[];objects.forEach((value,index)=>{offsets.push(Buffer.byteLength(result));result+=`${index+1} 0 obj\n${value}\nendobj\n`;});
 const xref=Buffer.byteLength(result);return result+`xref\n0 5\n0000000000 65535 f \n${offsets.map(offset=>String(offset).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size 5 /Root 1 0 R /Encrypt 4 0 R /ID [<${id.toString('hex')}> <${id.toString('hex')}>] >>\nstartxref\n${xref}\n%%EOF\n`;
}

test('real self-hosted PDF worker renders ordered PNG pages and closes on errors/cancel', {timeout:60000}, async () => {
 const page=await browser.newPage();const requests=[];
 page.on('request', request=>requests.push(request.url()));
 await page.goto(base+'/__pdf_import_test');
 const result=await page.evaluate(async ({source,encrypted})=>{
  const {importPdf}=await import('/src/lib/features/publishing/infrastructure/pdf-import.ts');
  const progress=[];
  const files=await importPdf(new File([source],'chapter.pdf'),{onProgress:(done,total)=>progress.push([done,total])});
  const colors=[];
  for(const file of files){const bitmap=await createImageBitmap(file);const canvas=document.createElement('canvas');canvas.width=1;canvas.height=1;const ctx=canvas.getContext('2d');ctx.drawImage(bitmap,0,0,1,1);colors.push([...ctx.getImageData(0,0,1,1).data]);bitmap.close();}
  let corrupt, cancelled, partialCancelled, timedOut, password;
  try{await importPdf(new File(['%PDF-1.7\nbroken'],'broken.pdf'));}catch(error){corrupt=error.code;}
  try{await importPdf(new File([encrypted],'locked.pdf'));}catch(error){password=error.code;}
  const controller=new AbortController();controller.abort();
  try{await importPdf(new File([source],'cancelled.pdf'),{signal:controller.signal});}catch(error){cancelled=error.code;}
  const partial=new AbortController();
  try{await importPdf(new File([source],'partial.pdf'),{signal:partial.signal,onProgress:done=>{if(done===1)partial.abort();}});}catch(error){partialCancelled=error.code;}
  const originalTimeout=window.setTimeout;
  window.setTimeout=(callback,ms,...args)=>originalTimeout(callback,ms===60000?1:ms,...args);
  try{await importPdf(new File([source],'timeout.pdf'));}catch(error){timedOut=error.code;}finally{window.setTimeout=originalTimeout;}
  // A failed/cancelled document must not poison a retry worker.
  const retried=await importPdf(new File([source],'retry.pdf'));
  return {names:files.map(file=>file.name), times:files.map(file=>file.lastModified),colors,progress,corrupt,cancelled,partialCancelled,timedOut,password,retried:retried.length};
 },{source:pdf(),encrypted:encryptedPdf()});
 assert.deepEqual(result.names,['pdf-page-001.png','pdf-page-002.png']);assert.deepEqual(result.times,[0,0]);
 assert.deepEqual(result.colors,[[255,0,0,255],[0,0,255,255]]);
 assert.deepEqual(result.progress,[[0,2],[1,2],[2,2]]);assert.equal(result.corrupt,'import_invalid');assert.equal(result.cancelled,'upload_cancelled');assert.equal(result.partialCancelled,'upload_cancelled');assert.equal(result.timedOut,'pdf_timeout');assert.equal(result.password,'pdf_password');assert.equal(result.retried,2);
 const license=await page.request.get(base+'/_pdf-assets/LICENSE');assert.equal(license.status(),200);assert.match(await license.text(),/Apache License/);
 const font=await page.request.get(base+'/_pdf-assets/standard_fonts/FoxitSerif.pfb');assert.equal(font.status(),200);
 assert.ok(requests.every(url=>url.startsWith(base)||url.startsWith('data:')), 'PDF import makes no external requests');
 await page.close();
});
