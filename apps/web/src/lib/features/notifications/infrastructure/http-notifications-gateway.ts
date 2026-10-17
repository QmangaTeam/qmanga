import type {NotificationsGateway,Notification} from '../application/notifications';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import type {AccountScope} from '../../identity/application/account-scope';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Деректер жарамсыз');return value as Record<string,unknown>;}
export class HttpNotificationsGateway implements NotificationsGateway{
 constructor(private api:ApiClient,private scope:AccountScope){}
 account(){return this.scope.currentId();}
 async list(accountId:string,after=''){const result=object(await this.api.call('/api/v1/notifications?after='+encodeURIComponent(after),'GET',undefined,accountId));if(!Array.isArray(result.items))throw new Error('Деректер жарамсыз');return {items:result.items.map(value=>{const item=object(value);return {id:String(item.id),kind:String(item.kind),title:String(item.title),body:String(item.body),releaseId:String(item.release_id),titleSlug:String(item.title_slug),projectId:typeof item.project_id==='string'?item.project_id:'',read:item.read===true,createdAt:String(item.created_at)} as Notification;}),hasMore:result.has_more===true};}
 async read(accountId:string,id:string){await this.api.call('/api/v1/notifications/'+encodeURIComponent(id)+'/read','PUT',{},accountId);}
}
