import type {GoogleLinkGateway,GoogleLinkStatus} from '../application/google-link';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {AppFailure} from '$lib/shared/domain/failure';
function object(value:unknown):value is Record<string,unknown>{return typeof value==='object'&&value!==null;}
export class HttpGoogleLinkGateway implements GoogleLinkGateway{
 constructor(private api:ApiClient){}
 async status(accountId:string):Promise<GoogleLinkStatus>{
  if(!accountId)throw new AppFailure('unauthorized');
  const value=await this.api.call('/api/v1/auth/google/link','GET',undefined,accountId);
  if(!object(value)||typeof value.enabled!=='boolean'||typeof value.linked!=='boolean')throw new AppFailure('unavailable');
  return {enabled:value.enabled,linked:value.linked};
 }
 async start(accountId:string):Promise<string>{
  if(!accountId)throw new AppFailure('unauthorized');
  const value=await this.api.call('/api/v1/auth/google/link/start','POST',{},accountId);
  if(!object(value)||typeof value.url!=='string')throw new AppFailure('unavailable');
  let url:URL;try{url=new URL(value.url);}catch{throw new AppFailure('unavailable');}
  if(url.origin!=='https://accounts.google.com'||url.username||url.password||url.pathname!=='/o/oauth2/v2/auth'||url.hash)throw new AppFailure('unavailable');
  return value.url;
 }
}
