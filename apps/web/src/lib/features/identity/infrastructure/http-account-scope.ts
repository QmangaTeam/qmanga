import type {AccountScope} from '../application/account-scope';
import {ApiClient,ApiFailure} from '$lib/shared/infrastructure/api-client';
export class HttpAccountScope implements AccountScope{
 constructor(private api:ApiClient){}
 async currentId(){try{const result=await this.api.call('/api/v1/auth/me');if(!result||typeof result!=='object'||!('user'in result)||!result.user||typeof result.user!=='object'||!('id'in result.user)||typeof result.user.id!=='string')throw new Error('Есептікжазба деректері жарамсыз');return result.user.id;}catch(failure){if(failure instanceof ApiFailure&&failure.status===401)return null;throw failure;}}
}
