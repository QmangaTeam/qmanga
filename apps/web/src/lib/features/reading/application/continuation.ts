import type {ReadingGateway,ProgressStorage} from './progress';

export interface Continuation {titleId:string;name:string;slug:string;releaseId:string}

export async function loadContinuation(gateway:ReadingGateway,storage:ProgressStorage,limit=4):Promise<Continuation[]> {
 if(limit<=0)return [];
 const account=await gateway.account();
 if(!account)return (await storage.list('guest')).filter(row=>row.progress.releaseId).slice(0,limit).map(row=>({...row,releaseId:row.progress.releaseId}));
 const rows:Continuation[]=[];
 const seen=new Set<string>();
 let cursor:string|undefined;
 while(rows.length<limit){
  const page=await gateway.list(account,cursor);
  for(const item of page.items){
   if(item.status==='reading'&&item.progress.releaseId&&!seen.has(item.titleId)){
    rows.push({titleId:item.titleId,name:item.name,slug:item.slug,releaseId:item.progress.releaseId});
    seen.add(item.titleId);
    if(rows.length===limit)return rows;
   }
  }
  const next=page.items.at(-1)?.titleId;
  if(!page.hasMore||!next||next===cursor)break;
  cursor=next;
 }
 return rows;
}
