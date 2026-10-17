import { MediaFailure, type Fragment, type ImageLoader, type ImageReference, type ImageTask, type Timer } from '../application/reader';
import {mediaNetwork} from '$lib/shared/infrastructure/media-network';
export class BrowserTimer implements Timer {later(callback:()=>void,milliseconds:number) {const timer=setTimeout(callback,milliseconds+Math.random()*200);return ()=>clearTimeout(timer);}}
export class BrowserImageLoader implements ImageLoader {
 constructor(private first?: HTMLImageElement) {}
 load(fragment: Fragment): ImageTask {
  const controller=new AbortController();let objectURL='';
  const first=this.first;if(first&&new URL(first.src).pathname===fragment.url){this.first=undefined;return {done:first.decode().then(()=>({url:fragment.url})).catch(()=>{throw new MediaFailure('Сурет жүктелмеді');}),cancel:()=>{first.removeAttribute('src');}};}
  const done=mediaNetwork.run(controller.signal,0,async()=>{
   const deadline=setTimeout(()=>controller.abort(),90000);
   try{
    const response=await fetch(fragment.url,{signal:controller.signal,credentials:'same-origin',cache:'no-cache'});
    if(!response.ok)throw new MediaFailure(response.status===404||response.status===410?'Бұл бет қолжетімсіз':'Сурет жүктелмеді',response.status===404||response.status===410,Number(response.headers.get('Retry-After'))*1000||0);
    const content=await response.arrayBuffer();
    if(content.byteLength!==fragment.bytes)throw new MediaFailure('Сурет толық жүктелмеді');
    const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',content)),byte=>byte.toString(16).padStart(2,'0')).join('');
    if(digest!==fragment.checksum)throw new MediaFailure('Суретті тексеру сәтсіз аяқталды');
    objectURL=URL.createObjectURL(new Blob([content],{type:fragment.mime}));const image=new Image();image.src=objectURL;await image.decode();
    if(controller.signal.aborted)throw new Error('Жүктеу тоқтатылды');return {url:objectURL};
   }catch(error){if(objectURL)URL.revokeObjectURL(objectURL);if(error instanceof MediaFailure)throw error;throw new MediaFailure('Қосылым үзілді немесе сурет толық жүктелмеді. Қайта көріңіз.');}finally{clearTimeout(deadline);}
  });
  return {done,cancel:()=>controller.abort()};
 }
 release(image: ImageReference) {if(image.url.startsWith('blob:'))URL.revokeObjectURL(image.url);}
}
