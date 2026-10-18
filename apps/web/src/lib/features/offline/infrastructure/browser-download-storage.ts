import {downloadFrom} from './download-record';
import {AppFailure} from '../../../shared/domain/failure';
import type { Download, DownloadStorage } from '../application/downloads';
import type { Fragment } from '../../reading/application/reader';
import { mediaNetwork } from '$lib/shared/infrastructure/media-network';
const cacheName='qmanga-saved-v1';
function database():Promise<IDBDatabase> {return new Promise((resolve,reject)=>{const request=indexedDB.open('qmanga-public-downloads',1);request.onupgradeneeded=()=>request.result.createObjectStore('downloads',{keyPath:'id'});request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(new AppFailure('storage_unavailable'));});}
async function transaction<Value>(mode:IDBTransactionMode,operation:(store:IDBObjectStore)=>IDBRequest<Value>):Promise<Value> {
 const connection=await database();return new Promise((resolve,reject)=>{const work=connection.transaction('downloads',mode),request=operation(work.objectStore('downloads'));let value:Value;request.onsuccess=()=>value=request.result;work.oncomplete=()=>{connection.close();resolve(value);};work.onabort=work.onerror=()=>{connection.close();reject(new AppFailure('storage_unavailable'));};});
}
async function valid(response:Response|undefined,fragment:Fragment) {if(!response)return false;const bytes=await response.arrayBuffer();if(bytes.byteLength!==fragment.bytes)return false;const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),byte=>byte.toString(16).padStart(2,'0')).join('');return digest===fragment.checksum;}
export class BrowserDownloadStorage implements DownloadStorage {
 list(){return transaction<Download[]>('readonly',store=>store.getAll()).then(items=>items.map(downloadFrom));}
 async write(download:Download){await transaction('readwrite',store=>store.put(download));}
 async remove(id:string) {const all=await this.list(),target=all.find(item=>item.id===id);if(!target)return;await transaction('readwrite',store=>store.delete(id));const cache=await caches.open(cacheName);const retained=new Set(all.filter(item=>item.id!==id).flatMap(item=>item.files));for(const url of target.files)if(!retained.has(url))await cache.delete(url);}
 async has(fragment:Fragment){return valid(await (await caches.open(cacheName)).match(fragment.url),fragment);}
 async store(fragment:Fragment,signal:AbortSignal) {
  await mediaNetwork.run(signal,10,async()=>{const response=await fetch(fragment.url,{signal,cache:'no-cache',credentials:'same-origin'}).catch(()=>{throw new AppFailure('network');});if(!response.ok)throw new AppFailure(response.status===404||response.status===410?'not_found':'download_failed');if(!await valid(response.clone(),fragment))throw new AppFailure('download_incomplete');try{await (await caches.open(cacheName)).put(fragment.url,response);}catch{throw new AppFailure('storage_unavailable');}});
 }
 async available(){const estimate=await navigator.storage?.estimate().catch(()=>null);return estimate?.quota?Math.max(0,estimate.quota-(estimate.usage||0)):null;}
 async install(){
  if(!('serviceWorker' in navigator))throw new AppFailure('offline_unsupported');
  const registration=await navigator.serviceWorker.register('/service-worker.js'),installing=registration.installing||registration.waiting;
  if(!registration.active&&installing&&installing.state!=='activated')await new Promise<void>((resolve,reject)=>{const deadline=setTimeout(()=>{installing.removeEventListener('statechange',changed);reject(new AppFailure('offline_install'));},90000);const changed=()=>{if(installing.state==='activated'||installing.state==='redundant'){clearTimeout(deadline);installing.removeEventListener('statechange',changed);if(installing.state==='activated')resolve();else reject(new AppFailure('offline_install'));}};installing.addEventListener('statechange',changed);changed();});
  await navigator.serviceWorker.ready;await navigator.storage?.persist?.().catch(()=>false);
 }
}
