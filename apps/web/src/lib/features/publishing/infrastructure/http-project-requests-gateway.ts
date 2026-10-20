import type {ProjectRequestsGateway} from '../application/project-requests';
import type {ProjectRequest,ProjectRequestInput,RequestState} from '../domain/project-request';
import type {ApiClient} from '$lib/shared/infrastructure/api-client';
import {AppFailure} from '$lib/shared/domain/failure';
function object(value:unknown):Record<string,unknown>{if(!value||typeof value!=='object')throw new AppFailure('unavailable');return value as Record<string,unknown>;}
function string(value:unknown){if(typeof value!=='string')throw new AppFailure('unavailable');return value;}
function request(value:unknown):ProjectRequest{
 const item=object(value),status=string(item.status);if(!['pending','approved','rejected','cancelled'].includes(status))throw new AppFailure('unavailable');
 return {coverURL:typeof item.cover_url==='string'?item.cover_url:undefined,id:string(item.id),teamId:string(item.team_id),teamName:string(item.team_name),titleId:string(item.title_id),titleName:string(item.title_name),name:string(item.name),originalName:string(item.original_name),slug:string(item.slug),kind:string(item.kind),description:string(item.description),sourceURL:string(item.source_url),note:string(item.note),status:status as RequestState,resolution:string(item.resolution),projectId:string(item.project_id),createdAt:string(item.created_at)};
}
export class HttpProjectRequestsGateway implements ProjectRequestsGateway {
 constructor(private api:ApiClient){}
 async list(admin=false,after='',filter={team:'',kind:''}){const value=object(await this.api.call('/api/v1/'+(admin?'admin':'studio')+'/project-requests?'+new URLSearchParams({after,...filter})));if(!Array.isArray(value.items))throw new AppFailure('unavailable');return {items:value.items.map(request),hasMore:value.has_more===true};}
 async create(input:ProjectRequestInput){await this.api.call('/api/v1/studio/project-requests','POST',{team_id:input.teamId,title_id:input.titleId,name:input.name,original_name:input.originalName,slug:input.slug,kind:input.kind,description:input.description,source_url:input.sourceURL,note:input.note});}
 async review(id:string,decision:'approve'|'reject',reason:string,titleId=''){await this.api.call('/api/v1/admin/project-requests/'+encodeURIComponent(id)+'/review','POST',{decision,reason,title_id:titleId});}
 async cancel(id:string){await this.api.call('/api/v1/studio/project-requests/'+encodeURIComponent(id)+'/cancel','POST',{});}
 async search(query:string){const value=object(await this.api.call('/api/v1/titles?q='+encodeURIComponent(query)+'&limit=24'));if(!Array.isArray(value.items))throw new AppFailure('unavailable');return value.items.map(value=>{const item=object(value);return {coverURL:typeof item.cover_url==='string'?item.cover_url:undefined,id:string(item.id),name:string(item.name),slug:string(item.slug)};});}
}
