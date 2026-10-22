import type {AccountPreferences,AccountProfilesGateway,PublicProfile} from '../application/account-profiles';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
function object(value:unknown):Record<string,unknown>{if(!value||typeof value!=='object')throw new Error('Invalid account response');return value as Record<string,unknown>;}
function preferences(value:unknown):AccountPreferences{const row=object(value);if(!Number.isSafeInteger(row.version)||Number(row.version)<0||typeof row.bio!=='string'||['public_profile','show_library','show_activity','allow_follow','notify_releases','notify_replies','notify_followers','email_digest'].some(key=>typeof row[key]!=='boolean'))throw new Error('Invalid account preferences');return row as unknown as AccountPreferences;}
function profile(value:unknown):PublicProfile{const row=object(value);if(['id','display_name','bio','avatar_url','created_at'].some(key=>typeof row[key]!=='string')||['public_profile','show_library','show_activity','allow_follow','is_owner','is_following'].some(key=>typeof row[key]!=='boolean')||['followers_count','following_count'].some(key=>!Number.isSafeInteger(row[key])||Number(row[key])<0)||!Array.isArray(row.library)||!Array.isArray(row.activity))throw new Error('Invalid profile');for(const value of row.library){const item=object(value);if(['title_id','slug','name','cover_url','status'].some(key=>typeof item[key]!=='string'))throw new Error('Invalid library');}for(const value of row.activity){const item=object(value);if(['label','href','created_at'].some(key=>typeof item[key]!=='string'))throw new Error('Invalid activity');}return row as unknown as PublicProfile;}
export class HttpAccountProfilesGateway implements AccountProfilesGateway {
 constructor(private api:ApiClient){}
 async preferences(accountId:string){return preferences(await this.api.call('/api/v1/account/preferences','GET',undefined,accountId));}
 async save(accountId:string,value:AccountPreferences){return preferences(await this.api.call('/api/v1/account/preferences','PUT',value,accountId));}
 async profile(id:string){return profile(await this.api.call('/api/v1/profiles/'+encodeURIComponent(id)));}
 async follow(accountId:string,id:string,following:boolean){await this.api.call('/api/v1/profiles/'+encodeURIComponent(id)+'/follow','PUT',{following},accountId);}
}
