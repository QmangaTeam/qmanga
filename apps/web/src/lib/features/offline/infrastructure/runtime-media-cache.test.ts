import {test,expect} from 'bun:test';
import {RuntimeMediaCache} from './runtime-media-cache';

function memoryCache() {
 const entries=new Map<string,Response>();
 const cache={match:async(key:string|Request)=>entries.get(typeof key==='string'?key:key.url)?.clone(),put:async(key:string|Request,response:Response)=>{entries.set(typeof key==='string'?key:key.url,response.clone());},delete:async(key:string|Request)=>entries.delete(typeof key==='string'?key:key.url),keys:async()=>[...entries.keys()].map(key=>new Request(key))};
 return {entries,cache};
}
function media(bytes:number,visibility='public, max-age=0, must-revalidate'){return new Response(new Uint8Array(bytes),{headers:{'content-type':'image/webp','content-length':String(bytes),'cache-control':visibility}});}
test('runtime cache evicts least recently used bytes without touching saved chapters',async()=>{
 const {cache,entries}=memoryCache();let clock=0;
 const runtime=new RuntimeMediaCache(cache,()=>++clock,10);
 await runtime.store('https://qmanga.test/api/v1/media/one',media(4));
 await runtime.store('https://qmanga.test/api/v1/media/two',media(4));
 expect(await runtime.read('https://qmanga.test/api/v1/media/one')).toBeDefined();
 await runtime.store('https://qmanga.test/api/v1/media/three',media(4));
 expect([...entries.keys()]).toEqual(['https://qmanga.test/api/v1/media/one','https://qmanga.test/api/v1/media/three']);
 expect((await runtime.read('https://qmanga.test/api/v1/media/three'))?.headers.get('content-length')).toBe('4');
});
test('runtime cache excludes private previews, partial responses, oversized and incomplete images',async()=>{
 const {cache,entries}=memoryCache();const runtime=new RuntimeMediaCache(cache,()=>1,10);
 const key='https://qmanga.test/api/v1/media/page';
 await runtime.store(key,media(4,'private, no-store'));
 await runtime.store(key,media(11));
 await runtime.store(key,new Response('html',{headers:{'cache-control':'public','content-type':'text/html'}}));
 await runtime.store(key,new Response('part',{status:206,headers:{'cache-control':'public','content-type':'image/webp','content-length':'4'}}));
 await runtime.store(key,new Response('short',{headers:{'cache-control':'public','content-type':'image/webp','content-length':'8'}}));
 expect(entries.size).toBe(0);
 await runtime.store(key,media(4));
 await runtime.remove(key);
 expect(await runtime.read(key)).toBeUndefined();
});
test('concurrent cache writes still respect the byte budget',async()=>{
 const {cache,entries}=memoryCache();let clock=0;const runtime=new RuntimeMediaCache(cache,()=>++clock,10);
 await Promise.all([1,2,3,4].map(index=>runtime.store('https://qmanga.test/api/v1/media/'+index,media(4))));
 expect(entries.size).toBe(2);
});
