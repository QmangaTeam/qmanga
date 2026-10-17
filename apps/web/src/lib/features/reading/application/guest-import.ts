import type {LocalProgress,Progress,ProgressStorage,ReadingGateway} from './progress';
export async function importGuest(gateway:ReadingGateway,storage:ProgressStorage,account:string,guest:LocalProgress,confirmedVersion?:number):Promise<Progress|null>{
 const cloud=await gateway.progress(account,guest.titleId);
 if(confirmedVersion===undefined&&cloud.version>0)return cloud;
 const result=await gateway.advance(account,{...storage.identity(),titleId:guest.titleId,releaseId:guest.progress.releaseId,revisionId:guest.progress.revisionId,pageId:guest.progress.pageId,fraction:guest.progress.fraction,baseVersion:confirmedVersion??cloud.version,completed:false});
 if(result.outcome==='conflict'||result.outcome==='stale')return result.progress;
 if(!await gateway.selection(account,guest.titleId))await gateway.save(account,{titleId:guest.titleId,status:'reading',favorite:false,preferredTeamId:'',notify:true});
 return null;
}
