import type {LibraryEntry,Progress,ProgressEvent,ReadingGateway,Selection} from '../application/progress';
import {ApiClient,ApiFailure} from '$lib/shared/infrastructure/api-client';
import type {AccountScope} from '../../identity/application/account-scope';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Деректер жарамсыз');return value as Record<string,unknown>;}
export function progressFrom(value:unknown):Progress {const item=object(value);return {titleId:String(item.title_id),releaseId:String(item.release_id),revisionId:String(item.revision_id),pageId:String(item.page_id),fraction:Number(item.fraction),version:Number(item.version),furthestOrder:Number(item.furthest_order),deviceId:String(item.device_id),sequence:Number(item.sequence)};}
export class HttpReadingGateway implements ReadingGateway {
 constructor(private api:ApiClient,private scope:AccountScope){}
 account(){return this.scope.currentId();}
 async selection(accountId:string,titleId:string){const result=object(await this.api.call('/api/v1/library/'+encodeURIComponent(titleId),'GET',undefined,accountId));if(!result.entry)return null;return this.entry(result.entry);}
 private entry(value:unknown):LibraryEntry{const item=object(value);return {titleId:String(item.title_id),name:String(item.name),slug:String(item.slug),status:String(item.status),favorite:item.favorite===true,preferredTeamId:String(item.preferred_team_id),notify:item.notify===true,unseen:Number(item.unseen),progress:progressFrom(item.progress)};}
 async progress(accountId:string,titleId:string){return progressFrom(await this.api.call('/api/v1/progress/'+encodeURIComponent(titleId),'GET',undefined,accountId));}
 async advance(accountId:string,event:ProgressEvent){
  const body={event_id:event.id,device_id:event.deviceId,sequence:event.sequence,base_version:event.baseVersion,release_id:event.releaseId,revision_id:event.revisionId,page_id:event.pageId,fraction:event.fraction,completed:event.completed};let value:unknown;
  try{value=await this.api.call('/api/v1/progress/'+encodeURIComponent(event.titleId),'POST',body,accountId);}catch(failure){if(failure instanceof ApiFailure&&failure.status===409&&object(failure.payload).progress)value=failure.payload;else throw failure;}
  const result=object(value);return {outcome:String(result.outcome),progress:progressFrom(result.progress)};
 }
 async list(accountId:string,after=''){const result=object(await this.api.call('/api/v1/library?after='+encodeURIComponent(after),'GET',undefined,accountId));if(!Array.isArray(result.items))throw new Error('Деректер жарамсыз');return {items:result.items.map(value=>{const item=object(value);return {titleId:String(item.title_id),name:String(item.name),slug:String(item.slug),status:String(item.status),favorite:item.favorite===true,preferredTeamId:String(item.preferred_team_id),notify:item.notify===true,unseen:Number(item.unseen),progress:progressFrom(item.progress)} as LibraryEntry;}),hasMore:result.has_more===true};}
 async save(accountId:string,input:Selection){await this.api.call('/api/v1/library/'+encodeURIComponent(input.titleId),'PUT',{status:input.status,favorite:input.favorite,preferred_team_id:input.preferredTeamId,notify:input.notify},accountId);}
 async remove(accountId:string,titleId:string){await this.api.call('/api/v1/library/'+encodeURIComponent(titleId),'DELETE',undefined,accountId);}
}
