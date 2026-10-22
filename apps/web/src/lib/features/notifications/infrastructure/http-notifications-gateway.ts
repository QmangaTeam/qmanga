import type {NotificationsGateway,Notification} from '../application/notifications';
import {notificationChanges} from './notification-changes';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import type {AccountScope} from '../../identity/application/account-scope';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Деректер жарамсыз');return value as Record<string,unknown>;}
export class HttpNotificationsGateway implements NotificationsGateway{
 constructor(private api:ApiClient,private scope:AccountScope){}
 watchAccountChange(changed:()=>void){const storage=(event:StorageEvent)=>{if(event.key==='qmanga:auth-epoch')changed();};window.addEventListener('storage',storage);window.addEventListener('qmanga:account-change',changed);return()=>{window.removeEventListener('storage',storage);window.removeEventListener('qmanga:account-change',changed);};}
 account(){return this.scope.currentId();}
 async unread(accountId:string){const value=object(await this.api.call('/api/v1/notifications/unread-count','GET',undefined,accountId));if(typeof value.count!=='number'||!Number.isSafeInteger(value.count)||value.count<0)throw new Error('Invalid count');return value.count;}
 async list(accountId:string,after='',unreadOnly=false,kind:''|'chapters'|'replies'=''){const result=object(await this.api.call('/api/v1/notifications?after='+encodeURIComponent(after)+(unreadOnly?'&unread=true':'')+(kind?'&kind='+kind:''),'GET',undefined,accountId));if(!Array.isArray(result.items))throw new Error('Деректер жарамсыз');return {items:result.items.map(value=>{const item=object(value);return {id:String(item.id),kind:String(item.kind),title:String(item.title),body:String(item.body),releaseId:String(item.release_id),titleSlug:String(item.title_slug),projectId:typeof item.project_id==='string'?item.project_id:'',commentId:typeof item.comment_id==='string'?item.comment_id:'',read:item.read===true,createdAt:String(item.created_at)} as Notification;}),hasMore:result.has_more===true};}
 async readAll(accountId:string){await this.api.call('/api/v1/notifications/read-all','PUT',{},accountId);notificationChanges.publish();}
 async read(accountId:string,id:string){await this.api.call('/api/v1/notifications/'+encodeURIComponent(id)+'/read','PUT',{},accountId);notificationChanges.publish();}
}
