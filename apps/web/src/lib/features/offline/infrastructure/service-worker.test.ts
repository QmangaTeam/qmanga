import {test,expect,mock} from 'bun:test';

test('offline shells are cached and updates wait for open readers to close',async()=>{
 const previous={self:globalThis.self,caches:globalThis.caches,fetch:globalThis.fetch};
 const handlers=new Map<string,(event:any)=>void>(),added:string[]=[],deleted:string[]=[];let skipped=0,claimed=0;
 const cache={addAll:async(paths:string[])=>{added.push(...paths);},match:async(path:string)=>['/saved','/offline-reader'].includes(path)?new Response(path):undefined};
 globalThis.self={location:{origin:'https://qmanga.test'},addEventListener:(type:string,handler:any)=>handlers.set(type,handler),skipWaiting:async()=>{skipped++;},clients:{claim:async()=>{claimed++;}}} as unknown as Window & typeof globalThis;
 globalThis.caches={open:async()=>cache,keys:async()=>['qmanga-static-old','qmanga-saved-v1'],delete:async(key:string)=>{deleted.push(key);return true;}} as unknown as CacheStorage;
 globalThis.fetch=(async()=>{throw new Error('offline');}) as typeof fetch;
 mock.module('$service-worker',()=>({build:['/app.js'],files:[],version:'new'}));
 try{
  await import('../../../../service-worker');
  let pending:Promise<unknown>=Promise.resolve();const waitUntil=(work:Promise<unknown>)=>{pending=work;};
  handlers.get('install')!({waitUntil});await pending;
  expect(added).toContain('/saved');expect(added).toContain('/offline-reader');expect(skipped).toBe(0);
  handlers.get('activate')!({waitUntil});await pending;
  expect(claimed).toBe(0);expect(deleted).toEqual(['qmanga-static-old']);
  const navigate=async(path:string)=>{let response:Promise<Response>|undefined;handlers.get('fetch')!({request:{url:'https://qmanga.test'+path,method:'GET',mode:'navigate',headers:new Headers()},respondWith:(value:Promise<Response>)=>{response=Promise.resolve(value);}});return response?await response:undefined;};
  expect(await (await navigate('/saved'))?.text()).toBe('/saved');
  expect((await navigate('/'))?.headers.get('Location')).toBe('https://qmanga.test/saved');
  expect((await navigate('/read/chapter-two'))?.headers.get('Location')).toBe('https://qmanga.test/offline-reader?release=chapter-two');
 }finally{globalThis.self=previous.self;globalThis.caches=previous.caches;globalThis.fetch=previous.fetch;mock.restore();}
});
