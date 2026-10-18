import type {ManagedAccount,ManagementGateway,AuditEntry} from '../application/management';
import type {AccountScope} from '../application/account-scope';
import type {ApiClient} from '$lib/shared/infrastructure/api-client';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Invalid response');return value as Record<string,unknown>;}
export class HttpManagementGateway implements ManagementGateway {
 constructor(private api:ApiClient,private scope:AccountScope){}
 account(){return this.scope.currentId();}
 async accounts(account:string,query:string,after=''){const response=object(await this.api.call('/api/v1/moderation/accounts?q='+encodeURIComponent(query)+'&after='+encodeURIComponent(after),'GET',undefined,account));if(!Array.isArray(response.items))throw new Error('Invalid accounts');return {items:response.items.map(value=>{const item=object(value);return {id:String(item.id),name:String(item.name),role:String(item.role),blocked:item.blocked===true,version:Number(item.version)};}),hasMore:response.has_more===true};}
 async change(account:string,target:ManagedAccount,reason:string){await this.api.call('/api/v1/moderation/accounts/'+encodeURIComponent(target.id),'PUT',{role:target.role,blocked:target.blocked,version:target.version,reason},account);}
 async audit(account:string,after=''){const response=object(await this.api.call('/api/v1/admin/audit?after='+encodeURIComponent(after),'GET',undefined,account));if(!Array.isArray(response.items))throw new Error('Invalid audit');return {items:response.items.map(value=>{const item=object(value);return {id:String(item.id),actor:String(item.actor),actorName:typeof item.actor_name==='string'?item.actor_name:'',targetName:typeof item.target_name==='string'?item.target_name:'',targetPath:typeof item.target_path==='string'&&/^\/(titles|teams|read|moderation|admin)\//.test(item.target_path)?item.target_path:'',before:safeState(item.before),after:safeState(item.after),action:String(item.action),entityType:String(item.entity_type),entityId:String(item.entity_id),reason:String(item.reason),createdAt:String(item.created_at)} as AuditEntry;}),hasMore:response.has_more===true};}
}

function safeState(value:unknown):Record<string,string|boolean|number>{
 if(!value||typeof value!=='object')return {};
 return Object.fromEntries(Object.entries(value).filter(([key,item])=>['role','blocked','hidden','status','revision_status','release_status','cover_asset_id'].includes(key)&&['string','boolean','number'].includes(typeof item))) as Record<string,string|boolean|number>;
}
