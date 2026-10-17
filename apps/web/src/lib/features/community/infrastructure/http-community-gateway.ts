import type {CommunityGateway,Comment,Discussion,Report,ModerationReport} from '../application/community';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import type {AccountScope} from '../../identity/application/account-scope';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Деректер жарамсыз');return value as Record<string,unknown>;}
function commentFrom(value:unknown):Comment{const item=object(value);return {id:String(item.id),userId:String(item.user_id),parentId:String(item.parent_id),body:String(item.body),spoiler:item.spoiler===true,hidden:item.hidden===true,createdAt:String(item.created_at),replies:Number(item.replies)};}
export class HttpCommunityGateway implements CommunityGateway {
 constructor(private api:ApiClient,private scope:AccountScope){}
 account(){return this.scope.currentId();}
 async replies(parentId:string,after=''){const result=object(await this.api.call(`/api/v1/community/comments/${encodeURIComponent(parentId)}/replies?after=${encodeURIComponent(after)}`));if(!Array.isArray(result.comments))throw new Error('Жауап деректері жарамсыз');return {comments:result.comments.map(commentFrom),next:String(result.next)};}
 async discussion(titleId:string,releaseId:string,after=''):Promise<Discussion>{const result=object(await this.api.call(`/api/v1/community/titles/${encodeURIComponent(titleId)}?release=${encodeURIComponent(releaseId)}&after=${encodeURIComponent(after)}`));if(!Array.isArray(result.comments))throw new Error('Пікірлер деректері жарамсыз');return {comments:result.comments.map(commentFrom),next:String(result.next),score:Number(result.score),ratings:Number(result.ratings)};}
 async comment(accountId:string,titleId:string,releaseId:string,parentId:string,body:string,spoiler:boolean){await this.api.call('/api/v1/community/comments','POST',{title_id:titleId,release_id:releaseId,parent_id:parentId,body,spoiler},accountId);}
 async edit(accountId:string,id:string,body:string,spoiler:boolean){await this.api.call('/api/v1/community/comments/'+encodeURIComponent(id),'PUT',{body,spoiler},accountId);}
 async remove(accountId:string,id:string){await this.api.call('/api/v1/community/comments/'+encodeURIComponent(id),'DELETE',undefined,accountId);}
 async rating(accountId:string,titleId:string,score:number){await this.api.call('/api/v1/community/titles/'+encodeURIComponent(titleId)+'/rating','PUT',{score},accountId);}
 async myRating(accountId:string,titleId:string){return Number(object(await this.api.call('/api/v1/community/titles/'+encodeURIComponent(titleId)+'/rating','GET',undefined,accountId)).score);}
 async report(accountId:string,input:Report){await this.api.call('/api/v1/community/reports','POST',{title_id:input.titleId,release_id:input.releaseId,page_id:input.pageId,target_type:input.targetType,target_id:input.targetId,category:input.category,reason:input.reason},accountId);}
 async reports(accountId:string,after=''){const result=object(await this.api.call('/api/v1/moderation/reports?after='+encodeURIComponent(after),'GET',undefined,accountId));if(!Array.isArray(result.items))throw new Error('Шағым деректері жарамсыз');return {items:result.items.map(value=>{const item=object(value);return {id:String(item.id),titleId:String(item.title_id),releaseId:String(item.release_id),pageId:String(item.page_id),targetType:String(item.target_type),targetId:String(item.target_id),category:String(item.category),reason:String(item.reason),status:String(item.status),createdAt:String(item.created_at)} as ModerationReport;}),hasMore:result.has_more===true};}
 async resolve(accountId:string,id:string,status:string,reason:string){await this.api.call('/api/v1/moderation/reports/'+encodeURIComponent(id),'PUT',{status,reason},accountId);}
 async visibility(accountId:string,id:string,hidden:boolean,reason:string){await this.api.call('/api/v1/moderation/comments/'+encodeURIComponent(id)+'/visibility','PUT',{hidden,reason},accountId);}
}
