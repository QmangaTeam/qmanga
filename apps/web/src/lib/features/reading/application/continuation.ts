import type {ReadingGateway,ProgressStorage,RecentReading,ReadReference} from './progress';

export type Continuation = RecentReading & {pageCount?:number;chapterProgress?:number};

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
  return {titleId:row.titleId,name:row.name,slug:row.slug,coverURL:row.coverURL||'',releaseId:row.progress.releaseId,revisionId:row.progress.revisionId,chapterLabel:row.chapterLabel||'',teamName:row.progress.teamName||'',...reference,progress:row.progress,updatedAt:row.updatedAt||0,pagePosition:row.pagePosition||1,availability:references===null?'last-known':!reference?'withdrawn':reference.revisionId!==row.progress.revisionId?'updated':'available'};
 });
}

// A public manifest supplies the denominator; never estimate it from saved position.
export async function continuationProgress(rows:Continuation[],gateway:Pick<import('./reader').ReaderGateway,'manifest'>):Promise<Continuation[]> {
 return Promise.all(rows.slice(0,4).map(async row=>{
  if(row.availability!=='available')return row;
  try{
   const manifest=await gateway.manifest(row.releaseId);
   if(manifest.releaseId!==row.releaseId||manifest.titleId!==row.titleId)return row;
   if(manifest.revisionId!==row.progress.revisionId)return {...row,availability:'updated' as const};
   const index=manifest.pages.findIndex(page=>page.id===row.progress.pageId);
   if(index<0||!manifest.pages.length||!Number.isFinite(row.progress.fraction))return row;
   const fraction=Math.max(0,Math.min(1,row.progress.fraction));
   return {...row,pagePosition:index+1,pageCount:manifest.pages.length,chapterProgress:(index+fraction)/manifest.pages.length};
  }catch{return row;}
 }));
}
