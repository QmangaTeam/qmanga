import {expect,test} from 'bun:test';
import {BrowserImageLoader} from './browser-image-loader';

test('verified image survives reader instance switch without another network request',async()=>{
 const originalFetch=globalThis.fetch,originalImage=globalThis.Image;
 const bytes=new Uint8Array([1,2,3]);const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),byte=>byte.toString(16).padStart(2,'0')).join('');
 let requests=0;
 globalThis.fetch=(async()=>{requests++;return new Response(bytes,{status:200});}) as typeof fetch;
 globalThis.Image=class {src='';async decode(){}} as unknown as typeof Image;
 try{
  const fragment={id:'cache-switch-test',url:'/api/v1/media/cache-switch-test',width:2,height:2,bytes:3,checksum:digest,mime:'image/webp',offset:0};
  const first=new BrowserImageLoader();const image=await first.load(fragment).done;first.release(image);
  const second=new BrowserImageLoader();const cached=second.cached(fragment);
  expect(cached?.url).toBe(image.url);expect(requests).toBe(1);
  second.release(cached!);
 }finally{globalThis.fetch=originalFetch;globalThis.Image=originalImage;}
});

test('private preview images are released instead of retained in shared cache',async()=>{
 const originalFetch=globalThis.fetch,originalImage=globalThis.Image;
 const bytes=new Uint8Array([4,5,6]);const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),byte=>byte.toString(16).padStart(2,'0')).join('');
 globalThis.fetch=(async()=>new Response(bytes,{status:200})) as typeof fetch;
 globalThis.Image=class {src='';async decode(){}} as unknown as typeof Image;
 try{
  const fragment={id:'private-preview-test',url:'/api/v1/media/private-preview-test',width:2,height:2,bytes:3,checksum:digest,mime:'image/webp',offset:0};
  const preview=new BrowserImageLoader(undefined,false);const image=await preview.load(fragment).done;preview.release(image);
  expect(new BrowserImageLoader().cached(fragment)).toBeUndefined();
 }finally{globalThis.fetch=originalFetch;globalThis.Image=originalImage;}
});
