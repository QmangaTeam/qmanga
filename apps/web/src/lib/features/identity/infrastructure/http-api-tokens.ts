import type {ApiToken,ApiTokensGateway} from '../application/api-tokens';
import type {ApiClient} from '$lib/shared/infrastructure/api-client';
function object(value:unknown):Record<string,unknown>{if(!value||typeof value!=='object')throw new Error('Invalid token response');return value as Record<string,unknown>;}
function string(value:unknown){if(typeof value!=='string')throw new Error('Invalid token metadata');return value;}
function metadata(value:unknown):ApiToken{const item=object(value);return {id:string(item.id),name:string(item.name),prefix:string(item.prefix),createdAt:string(item.created_at),expiresAt:string(item.expires_at),lastUsedAt:typeof item.last_used_at==='string'?item.last_used_at:null,revokedAt:typeof item.revoked_at==='string'?item.revoked_at:null,directPublish:item.direct_publish===true};}
export class HttpApiTokens implements ApiTokensGateway {
 constructor(private api:ApiClient){}
 async list(account:string){const response=object(await this.api.call('/api/v1/admin/api-tokens','GET',undefined,account));if(!Array.isArray(response.items))throw new Error('Invalid token list');return response.items.map(metadata);}
 async create(account:string,input:{name:string;expiresInDays:number;directPublish:boolean}){const response=object(await this.api.call('/api/v1/admin/api-tokens','POST',{name:input.name,expires_in_days:input.expiresInDays,direct_publish:input.directPublish},account));return {token:string(response.token),item:metadata(response.item)};}
 async revoke(account:string,id:string){await this.api.call('/api/v1/admin/api-tokens/'+encodeURIComponent(id),'DELETE',undefined,account);}
}
