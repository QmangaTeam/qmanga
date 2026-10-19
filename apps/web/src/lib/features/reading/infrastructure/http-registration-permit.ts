import type {RegistrationPermit} from '../application/registration-import';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
export class HttpRegistrationPermit implements RegistrationPermit{
 constructor(private api:ApiClient){}
 async allowed(account:string){const result=await this.api.call('/api/v1/auth/guest-import','GET',undefined,account);if(!result||typeof result!=='object'||!('allowed'in result)||typeof result.allowed!=='boolean')throw Error('Invalid registration permit');return result.allowed;}
 async complete(account:string){await this.api.call('/api/v1/auth/guest-import/complete','POST',{},account);}
}
