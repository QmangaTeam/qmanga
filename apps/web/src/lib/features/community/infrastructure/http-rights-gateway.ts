import type {RightsGateway,RightsInput} from '../application/rights';
import {ApiFailure} from '$lib/shared/infrastructure/api-client';
export class HttpRightsGateway implements RightsGateway {
 constructor(private request:typeof fetch){}
 private async call(path:string,input:unknown):Promise<Record<string,unknown>> {
  let response:Response;
  try{response=await this.request('/api/v1/rights-reports'+path,{method:'POST',credentials:'omit',headers:{'Content-Type':'application/json'},body:JSON.stringify(input),signal:AbortSignal.timeout(20000)});}catch{throw new ApiFailure(0,'network');}
  let value:unknown;try{value=await response.json();}catch{throw new ApiFailure(503,'unavailable');}
  if(!response.ok)throw new ApiFailure(response.status,'request_failed',value);
  if(!value||typeof value!=='object')throw new ApiFailure(503,'unavailable');
  return value as Record<string,unknown>;
 }
 async create(input:RightsInput){const result=await this.call('',{title_slug:input.titleSlug,email:input.email,reason:input.reason});return {id:String(result.id),token:String(result.token)};}
 async status(id:string,token:string){const result=await this.call('/status',{id,token});return {status:String(result.status),resolution:String(result.resolution)};}
}
