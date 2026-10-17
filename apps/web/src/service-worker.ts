/// <reference lib="webworker" />
import {build,files,version} from '$service-worker';
const worker=self as unknown as ServiceWorkerGlobalScope;
const staticCache='qmanga-static-'+version;
worker.addEventListener('install',event=>{event.waitUntil((async()=>{const cache=await caches.open(staticCache);await cache.addAll([...build,...files,'/offline-reader']);await worker.skipWaiting();})());});
worker.addEventListener('activate',event=>{event.waitUntil((async()=>{for(const key of await caches.keys())if(key.startsWith('qmanga-static-')&&key!==staticCache)await caches.delete(key);await worker.clients.claim();})());});
worker.addEventListener('fetch',event=>{
 const request=event.request,url=new URL(request.url);if(request.method!=='GET'||url.origin!==worker.location.origin)return;
 if(url.pathname==='/offline-reader'){event.respondWith(caches.open(staticCache).then(async cache=>(await cache.match('/offline-reader'))||fetch(request)));return;}
 if(build.includes(url.pathname)||files.includes(url.pathname)){event.respondWith(caches.open(staticCache).then(async cache=>(await cache.match(url.pathname))||fetch(request)));return;}
 if(/^\/api\/v1\/media\/[a-zA-Z0-9-]+$/.test(url.pathname)){event.respondWith(fetch(request).catch(async()=>{const saved=await (await caches.open('qmanga-saved-v1')).match(url.pathname);return saved||new Response('Желі жоқ',{status:503});}));return;}
 if(request.mode==='navigate'&&/^\/read\/[^/]+$/.test(url.pathname)&&!url.searchParams.has('preview')){event.respondWith(fetch(request).catch(()=>Promise.resolve(Response.redirect(new URL('/offline-reader?release='+encodeURIComponent(url.pathname.split('/')[2]),worker.location.origin),302))));}
});
