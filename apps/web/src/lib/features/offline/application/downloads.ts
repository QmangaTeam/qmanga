import {AppFailure} from '../../../shared/domain/failure';
import type { Fragment, Manifest, Tier } from '../../reading/application/reader';
export interface Download {id:string;manifest:Manifest;tier:Tier;status:'partial'|'ready'|'error';files:string[];bytes:number;total:number;message:string;notice?:string;updatedAt:number}
export interface DownloadStorage {
 list():Promise<Download[]>;
 write(download:Download):Promise<void>;
 remove(id:string):Promise<void>;
 has(fragment:Fragment):Promise<boolean>;
 store(fragment:Fragment,signal:AbortSignal):Promise<void>;
 available():Promise<number|null>;
 install():Promise<void>;
}
export function expectedFiles(manifest:Manifest,tier:Tier):Fragment[] {return manifest.pages.flatMap(page=>page.variants[tier]);}
export class Downloads {
 constructor(private storage:DownloadStorage,private now:()=>number){}
 list(){return this.storage.list();}
 remove(id:string){return this.storage.remove(id);}
 async save(manifest:Manifest,tier:Tier,signal:AbortSignal,changed:(download:Download)=>void):Promise<Download> {
  const files=expectedFiles(manifest,tier),id=[manifest.releaseId,manifest.revisionId,tier].join(':'),total=files.reduce((sum,file)=>sum+file.bytes,0),existing=await this.storage.list(),current=existing.find(item=>item.id===id);
  const used=existing.reduce((sum,item)=>sum+item.bytes,0),free=await this.storage.available();
  const ceiling=Math.min(200*1024*1024,free===null?200*1024*1024:used+Math.floor(free*.2));
  if(used-(current?.bytes||0)+total>ceiling)throw new AppFailure('storage_full');
  const download:Download={id,manifest,tier,status:'partial',files:[],bytes:0,total,message:'',notice:'download_partial',updatedAt:this.now()};
  const persist=async()=>{download.updatedAt=this.now();await this.storage.write({...download,files:[...download.files]});changed({...download,files:[...download.files]});};
  await persist();
  try{
   await this.storage.install();
   for(const fragment of files){if(signal.aborted)throw new AppFailure('download_cancelled');if(!await this.storage.has(fragment))await this.storage.store(fragment,signal);if(!await this.storage.has(fragment))throw new AppFailure('download_incomplete');download.files.push(fragment.url);download.bytes+=fragment.bytes;await persist();}
   download.status='ready';download.notice='none';await persist();return download;
  }catch(failure){download.status=signal.aborted?'partial':'error';download.notice=signal.aborted?'download_partial':'download_failed';download.message='';await persist();throw failure;}
 }
 async reconcile(download:Download):Promise<Download> {
  const files=[];let bytes=0;for(const fragment of expectedFiles(download.manifest,download.tier)){if(await this.storage.has(fragment)){files.push(fragment.url);bytes+=fragment.bytes;}}
  const value={...download,files,bytes,status:files.length===expectedFiles(download.manifest,download.tier).length?'ready' as const:'partial' as const};await this.storage.write(value);return value;
 }
}
