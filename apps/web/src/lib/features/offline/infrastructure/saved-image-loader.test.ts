import {test,expect} from 'bun:test';
import {SavedImageLoader} from './saved-image-loader';
import type {Fragment} from '../../reading/application/reader';
test('explicit offline reading uses saved bytes without any network attempt',async()=>{
 const originalFetch=globalThis.fetch,originalCaches=globalThis.caches,originalImage=globalThis.Image;let requests=0;
 const content=new Uint8Array([1,2,3]),hash=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',content)),byte=>byte.toString(16).padStart(2,'0')).join('');
 globalThis.fetch=(()=>{requests++;throw new Error('server is unavailable');}) as unknown as typeof fetch;
 globalThis.caches={open:async()=>({match:async()=>new Response(content)})} as unknown as CacheStorage;
 globalThis.Image=class{src='';decode(){return Promise.resolve();}} as unknown as typeof Image;
 try{const loader=new SavedImageLoader(),reference=await loader.load({url:'/api/v1/media/asset',bytes:3,checksum:hash,mime:'image/webp'} as Fragment).done;expect(reference.url.startsWith('blob:')).toBe(true);expect(requests).toBe(0);loader.release(reference);}finally{globalThis.fetch=originalFetch;globalThis.caches=originalCaches;globalThis.Image=originalImage;}
});
