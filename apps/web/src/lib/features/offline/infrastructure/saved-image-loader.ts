import {MediaFailure,type Fragment,type ImageLoader,type ImageReference,type ImageTask} from '../../reading/application/reader';
export class SavedImageLoader implements ImageLoader{
 load(fragment:Fragment):ImageTask{
  let cancelled=false,objectURL='';
  const done=(async()=>{try{const response=await(await caches.open('qmanga-saved-v1')).match(fragment.url);if(!response)throw new MediaFailure('Браузер бұл суретті өшірген. Желіге қосылып, тарауды қайта сақтаңыз.',true);const content=await response.arrayBuffer();const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',content)),byte=>byte.toString(16).padStart(2,'0')).join('');if(content.byteLength!==fragment.bytes||digest!==fragment.checksum)throw new MediaFailure('Сақталған сурет толық емес. Тарауды қайта сақтаңыз.',true);if(cancelled)throw new Error('Оқу тоқтатылды');objectURL=URL.createObjectURL(new Blob([content],{type:fragment.mime}));const image=new Image();image.src=objectURL;await image.decode();if(cancelled)throw new Error('Оқу тоқтатылды');return {url:objectURL};}catch(failure){if(objectURL)URL.revokeObjectURL(objectURL);throw failure;}})();
  return {done,cancel:()=>{cancelled=true;}};
 }
 release(reference:ImageReference){URL.revokeObjectURL(reference.url);}
}
