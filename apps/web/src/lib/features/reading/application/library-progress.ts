import type {LibraryEntry} from './progress';
import type {Manifest,ReaderGateway} from './reader';
export interface LibraryPosition {page:number;pages:number;fraction:number;chapter:string}
export class LibraryProgressReader {
 private active=0;
 private waiting:(()=>void)[]=[];
 private manifests=new Map<string,Promise<Manifest>>();
 constructor(private gateway:Pick<ReaderGateway,'manifest'>){}
 private manifest(id:string){
  let result=this.manifests.get(id);
  if(!result){result=(async()=>{if(this.active>=4)await new Promise<void>(resolve=>this.waiting.push(resolve));else this.active++;
   try{return await this.gateway.manifest(id);}finally{const next=this.waiting.shift();if(next)next();else this.active--;}
  })();this.manifests.set(id,result);result.catch(()=>{this.manifests.delete(id);});}
  return result;
 }
 async read(entry:LibraryEntry):Promise<LibraryPosition|null>{
  const progress=entry.progress;if(!progress.releaseId)return null;
  try{const manifest=await this.manifest(progress.releaseId);
   if(manifest.titleId!==entry.titleId||manifest.releaseId!==progress.releaseId||manifest.revisionId!==progress.revisionId)return null;
   const index=manifest.pages.findIndex(page=>page.id===progress.pageId);
   if(index<0||!Number.isFinite(progress.fraction))return null;
   return {page:index+1,pages:manifest.pages.length,fraction:(index+Math.max(0,Math.min(1,progress.fraction)))/manifest.pages.length,chapter:manifest.chapterLabel};
  }catch{return null;}
 }
}
