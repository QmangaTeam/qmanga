import type {ReadingGateway,ProgressStorage,RecentReading,ReadReference} from './progress';

export type Continuation = RecentReading;

export async function loadContinuation(gateway:ReadingGateway,storage:ProgressStorage,limit=4):Promise<Continuation[]> {
 limit=Math.min(12,Math.max(0,Math.floor(limit)));if(!limit)return [];
 const account=await gateway.account();
 if(account)return gateway.recent(account,limit);
 const local=(await storage.recent('guest',limit)).filter(row=>row.progress.releaseId);
 if(!local.length)return [];
 let references:ReadReference[]|null=null;
 try{references=await gateway.references([...new Set(local.map(row=>row.progress.releaseId))]);}catch{}
 return local.map(row=>{
  const reference=references?.find(item=>item.releaseId===row.progress.releaseId&&item.titleId===row.titleId);
  return {titleId:row.titleId,name:row.name,slug:row.slug,coverURL:row.coverURL||'',releaseId:row.progress.releaseId,revisionId:row.progress.revisionId,chapterLabel:row.chapterLabel||'',...reference,progress:row.progress,updatedAt:row.updatedAt||0,pagePosition:row.pagePosition||1,availability:references===null?'last-known':!reference?'withdrawn':reference.revisionId!==row.progress.revisionId?'updated':'available'};
 });
}
