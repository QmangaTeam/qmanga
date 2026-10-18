import type {CommunityGateway,Comment,Discussion,Report,ModerationReport} from '../application/community';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import type {AccountScope} from '../../identity/application/account-scope';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Деректер жарамсыз');return value as Record<string,unknown>;}
function commentFrom(value:unknown):Comment{const item=object(value);return {displayName:typeof item.display_name==='string'?item.display_name:'',id:String(item.id),userId:String(item.user_id),parentId:String(item.parent_id),body:String(item.body),spoiler:item.spoiler===true,hidden:item.hidden===true,createdAt:String(item.created_at),replies:Number(item.replies)};}
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
 async myReports(accountId:string,after=''){const result=object(await this.api.call('/api/v1/community/reports?after='+encodeURIComponent(after),'GET',undefined,accountId));if(!Array.isArray(result.items))throw new Error('Invalid reports');return {items:result.items.map(value=>{const item=object(value);return {id:String(item.id),category:String(item.category),reason:String(item.reason),status:String(item.status),resolution:String(item.resolution),createdAt:String(item.created_at)};}),hasMore:result.has_more===true};}
 async reports(accountId:string,after=''){
  const result=object(await this.api.call('/api/v1/moderation/reports?after='+encodeURIComponent(after),'GET',undefined,accountId));
  if(!Array.isArray(result.items))throw new Error('Шағым деректері жарамсыз');
  const value=(item:Record<string,unknown>,key:string)=>typeof item[key]==='string'?item[key] as string:'';
  return {items:result.items.map(raw=>{const item=object(raw);return {
   contactEmail:value(item,'contact_email'),id:value(item,'id'),titleId:value(item,'title_id'),releaseId:value(item,'release_id'),pageId:value(item,'page_id'),targetType:value(item,'target_type'),targetId:value(item,'target_id'),category:value(item,'category'),reason:value(item,'reason'),status:value(item,'status'),createdAt:value(item,'created_at'),
   titleName:value(item,'title_name'),titleSlug:value(item,'title_slug'),teamName:value(item,'team_name'),teamSlug:value(item,'team_slug'),chapterLabel:value(item,'chapter_label'),releaseStatus:value(item,'release_status'),revisionId:value(item,'revision_id'),pagePosition:Number(item.page_position)||0,commentBody:value(item,'comment_body'),commentAuthor:value(item,'comment_author'),commentHidden:item.comment_hidden===true,commentDeleted:item.comment_deleted===true,
   history:Array.isArray(item.history)?item.history.map(raw=>{const decision=object(raw);return {action:value(decision,'action'),status:value(decision,'status'),actorName:value(decision,'actor_name'),reason:value(decision,'reason'),createdAt:value(decision,'created_at')};}):[]
  } as ModerationReport;}),hasMore:result.has_more===true,next:value(result,'next')};
 }
 async resolve(accountId:string,id:string,status:string,reason:string){await this.api.call('/api/v1/moderation/reports/'+encodeURIComponent(id),'PUT',{status,reason},accountId);}
 async visibility(accountId:string,id:string,hidden:boolean,reason:string){await this.api.call('/api/v1/moderation/comments/'+encodeURIComponent(id)+'/visibility','PUT',{hidden,reason},accountId);}
}
