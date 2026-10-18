import { MediaFailure, type Fragment, type ImageLoader, type ImageReference, type ImageTask, type Timer } from '../application/reader';
import {mediaNetwork} from '$lib/shared/infrastructure/media-network';
type CachedImage={url:string;bytes:number;references:number};
const verified=new Map<string,CachedImage>();
const verifiedBudget=32*1024*1024;
let verifiedBytes=0;
const cacheKey=(fragment:Fragment)=>fragment.id+':'+fragment.checksum;
function remember(fragment:Fragment,url:string):boolean{
 const key=cacheKey(fragment);if(fragment.bytes>verifiedBudget||verified.has(key))return false;
 for(const [oldKey,entry] of verified){if(verifiedBytes+fragment.bytes<=verifiedBudget)break;if(entry.references)continue;verified.delete(oldKey);verifiedBytes-=entry.bytes;URL.revokeObjectURL(entry.url);}
 if(verifiedBytes+fragment.bytes>verifiedBudget)return false;
 verified.set(key,{url,bytes:fragment.bytes,references:1});verifiedBytes+=fragment.bytes;return true;
}
export class BrowserTimer implements Timer {later(callback:()=>void,milliseconds:number) {const timer=setTimeout(callback,milliseconds+Math.random()*200);return ()=>clearTimeout(timer);}}
export class BrowserImageLoader implements ImageLoader {
 constructor(private first?: HTMLImageElement,private cacheable=true) {}
 cached(fragment:Fragment):ImageReference|undefined{if(!this.cacheable)return;const key=cacheKey(fragment),entry=verified.get(key);if(!entry)return;entry.references++;verified.delete(key);verified.set(key,entry);return {url:entry.url};}
 load(fragment: Fragment,priority=0): ImageTask {
  const controller=new AbortController();let objectURL='';
  const cached=this.cached(fragment);if(cached)return {done:Promise.resolve(cached),cancel:()=>{}};
  const first=this.first;if(first&&new URL(first.src).pathname===fragment.url){this.first=undefined;return {done:first.decode().then(()=>({url:fragment.url})).catch(()=>{throw new MediaFailure('Сурет жүктелмеді');}),cancel:()=>{first.removeAttribute('src');}};}
  const done=mediaNetwork.run(controller.signal,priority,async()=>{
   const deadline=setTimeout(()=>controller.abort(),90000);
   try{
    const response=await fetch(fragment.url,{signal:controller.signal,credentials:'same-origin',cache:'default'});
    if(!response.ok)throw new MediaFailure(response.status===404||response.status===410?'Бұл бет қолжетімсіз':'Сурет жүктелмеді',response.status===404||response.status===410,Number(response.headers.get('Retry-After'))*1000||0);
    const content=await response.arrayBuffer();
    if(content.byteLength!==fragment.bytes)throw new MediaFailure('Сурет толық жүктелмеді');
    const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',content)),byte=>byte.toString(16).padStart(2,'0')).join('');
    if(digest!==fragment.checksum)throw new MediaFailure('Суретті тексеру сәтсіз аяқталды');
    objectURL=URL.createObjectURL(new Blob([content],{type:fragment.mime}));const image=new Image();image.src=objectURL;await image.decode();
    if(controller.signal.aborted)throw new Error('Жүктеу тоқтатылды');if(this.cacheable)remember(fragment,objectURL);return {url:objectURL};
   }catch(error){if(objectURL)URL.revokeObjectURL(objectURL);if(error instanceof MediaFailure)throw error;throw new MediaFailure('Қосылым үзілді немесе сурет толық жүктелмеді. Қайта көріңіз.');}finally{clearTimeout(deadline);}
  });
  return {done,cancel:()=>controller.abort()};
 }
 release(image: ImageReference) {if(!image.url.startsWith('blob:'))return;for(const entry of verified.values())if(entry.url===image.url){entry.references=Math.max(0,entry.references-1);return;}URL.revokeObjectURL(image.url);}
}
