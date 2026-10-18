import type {AvatarSource,AvatarGateway} from '../application/avatars';
import {AppFailure} from '$lib/shared/domain/failure';
export class HttpAvatarGateway implements AvatarGateway {
 constructor(private request:typeof fetch){}
 async upload(kind:'user'|'team',ownerId:string,source:AvatarSource):Promise<void>{
  if(!(source instanceof Blob))throw new AppFailure('validation');
  const csrfResponse=await this.request('/api/v1/auth/csrf',{credentials:'same-origin',signal:AbortSignal.timeout(10000)});
  if(!csrfResponse.ok)throw new AppFailure('unauthorized');
  const csrf:unknown=await csrfResponse.json();if(typeof csrf!=='object'||csrf===null||!('csrf_token'in csrf)||typeof csrf.csrf_token!=='string')throw new AppFailure('unavailable');
  const response=await this.request('/api/v1/avatars/'+(kind==='user'?'user':'team/'+encodeURIComponent(ownerId)),{method:'POST',credentials:'same-origin',headers:{'Content-Type':source.type,'X-CSRF-Token':csrf.csrf_token},body:source,signal:AbortSignal.timeout(30000)});
  if(!response.ok)throw new AppFailure(response.status===403?'forbidden':response.status===400||response.status===413?'validation':'unavailable');
 }
}
