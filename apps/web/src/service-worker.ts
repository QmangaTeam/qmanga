/// <reference lib="webworker" />
import {build,files,version} from '$service-worker';
import {RuntimeMediaCache} from './lib/features/offline/infrastructure/runtime-media-cache';
const worker=self as unknown as ServiceWorkerGlobalScope;
const staticCache='qmanga-static-'+version;
const runtimeMedia=caches.open('qmanga-runtime-media-v1').then(cache=>new RuntimeMediaCache(cache,()=>Date.now()));
// Let an update wait until existing tabs close: their reader still needs its own build assets.
worker.addEventListener('install',event=>{event.waitUntil((async()=>{const cache=await caches.open(staticCache);await cache.addAll([...build,...files,'/offline-reader','/saved']);})());});
worker.addEventListener('activate',event=>{event.waitUntil((async()=>{for(const key of await caches.keys())if(key.startsWith('qmanga-static-')&&key!==staticCache)await caches.delete(key);})());});
worker.addEventListener('fetch',event=>{
 const request=event.request,url=new URL(request.url);if(request.method!=='GET'||url.origin!==worker.location.origin)return;
 if(url.pathname==='/offline-reader'||url.pathname==='/saved'){event.respondWith(caches.open(staticCache).then(async cache=>(await cache.match(url.pathname))||fetch(request)));return;}
 if(build.includes(url.pathname)||files.includes(url.pathname)){event.respondWith(caches.open(staticCache).then(async cache=>(await cache.match(url.pathname))||fetch(request)));return;}
 if(/^\/api\/v1\/media\/[a-zA-Z0-9-]+$/.test(url.pathname)&&!url.search&&!request.headers.has('Range')){
  event.respondWith((async()=>{
   let response:Response;
   try{response=await fetch(request);}catch(failure){
    if(request.signal.aborted)throw failure;
    const saved=await caches.open('qmanga-saved-v1').then(cache=>cache.match(url.pathname)).catch(()=>undefined);
    const recent=await runtimeMedia.then(cache=>cache.read(url.href)).catch(()=>undefined);
    return saved||recent||new Response('Желі жоқ',{status:503});
   }
   if([401,403,404,410].includes(response.status)){
    event.waitUntil(Promise.all([runtimeMedia.then(cache=>cache.remove(url.href)),caches.open('qmanga-saved-v1').then(cache=>cache.delete(url.pathname))]).catch(()=>{}));
   }else if(response.ok){const copy=response.clone();event.waitUntil(runtimeMedia.then(cache=>cache.store(url.href,copy)).catch(()=>{}));}
   return response;
  })());return;
 }
 if(request.mode==='navigate'){event.respondWith(fetch(request).catch(()=>{
  const target=/^\/read\/[^/]+$/.test(url.pathname)&&!url.searchParams.has('preview')?'/offline-reader?release='+encodeURIComponent(url.pathname.split('/')[2]):'/saved';
  return Response.redirect(new URL(target,worker.location.origin),302);
 }));}
});
