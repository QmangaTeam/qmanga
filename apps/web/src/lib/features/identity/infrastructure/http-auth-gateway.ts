import type { AuthGateway, AuthResult } from '../application/auth';
import type { Account, Credentials } from '../domain/account';
function object(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null; }
function accountView(value: unknown): Account | undefined {
 if (!object(value) || typeof value.id !== 'string' || typeof value.email !== 'string' || typeof value.email_verified !== 'boolean') return;
 if (value.role !== 'reader' && value.role !== 'moderator' && value.role !== 'admin') return;
 return { avatarURL:typeof value.avatar_url==='string'&&/^\/media\/avatars\/user\/[a-zA-Z0-9-]+$/.test(value.avatar_url)?value.avatar_url:undefined, displayName:typeof value.display_name==='string'?value.display_name:'',id: value.id, email: value.email, role: value.role, verified: value.email_verified };
}
export class HttpAuthGateway implements AuthGateway {
 constructor(private base: string, private request: typeof fetch, private cookie = '',private accountChanged:()=>Promise<void>=async()=>{}) {}
 private async call(action: string, input?: Record<string, string>, csrf?: string): Promise<AuthResult> {
  try {
   const response = await this.request(this.base + '/api/v1/auth/' + action, {
    method: input === undefined ? 'GET' : 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...(this.cookie ? { cookie: this.cookie } : {}), ...(csrf ? { 'X-CSRF-Token': csrf } : {}) },
    ...(input === undefined ? {} : { body: JSON.stringify(input) }),
    signal: AbortSignal.timeout(15000)
   });
   if (response.status === 204) return { ok: true, message: 'Сәтті шықтыңыз' };
   const payload: unknown = await response.json();
   if (!object(payload)) throw new Error('Invalid response');
   if (!response.ok) return { ok: false, code: object(payload.error) && typeof payload.error.code==='string' ? payload.error.code : 'unavailable', message: object(payload.error) && typeof payload.error.message === 'string' ? payload.error.message : 'Сұрауды орындау мүмкін болмады' };
   return { ok: true, message: typeof payload.message === 'string' ? payload.message : '', account: accountView(payload.user) };
  } catch { return { ok: false, code:'network', message: 'Қосылым үзілді. Қайталап көріңіз.' }; }
 }
 async completeExternalLogin(){await this.accountChanged();}
 async googleStatus(){try{const response=await this.request(this.base+'/api/v1/auth/google/status',{signal:AbortSignal.timeout(10000)});const data:unknown=await response.json();return response.ok&&object(data)&&data.enabled===true;}catch{return false;}}
 async googleStart(returnPath:string,nickname:string){try{const response=await this.request(this.base+'/api/v1/auth/google/start',{method:'POST',credentials:'same-origin',headers:{'Content-Type':'application/json'},body:JSON.stringify({return_path:returnPath,nickname}),signal:AbortSignal.timeout(15000)});const data:unknown=await response.json();if(response.ok&&object(data)&&typeof data.url==='string'&&data.url.startsWith('https://accounts.google.com/o/oauth2/v2/auth?'))return {url:data.url};return {code:'validation'};}catch{return {code:'network'};}}
 async login(input: Credentials) {const result=await this.call('login', { email: input.email, password: input.password });if(result.ok)await this.accountChanged();return result;}
 register(input: Credentials) { return this.call('register', { email: input.email, password: input.password, display_name: input.displayName || '' }); }
 async current() { const result = await this.call('me'); return result.ok ? result.account || null : null; }
 requestReset(email: string) { return this.call('forgot-password', { email }); }
 resend(email: string) { return this.call('resend-verification', { email }); }
 verify(token: string) { return this.call('verify-email', { token }); }
 reset(token: string, password: string) { return this.call('reset-password', { token, password }); }
 private async authenticated(action:string,input:Record<string,string>): Promise<AuthResult> {
  try {
   const response = await this.request(this.base + '/api/v1/auth/csrf', { credentials: 'same-origin', signal: AbortSignal.timeout(10000) });
   const payload: unknown = await response.json();
   if (!response.ok || !object(payload) || typeof payload.csrf_token !== 'string') return { ok: false, code:'unauthorized', message: 'Қайта кіріңіз' };
   return await this.call(action,input,payload.csrf_token);
  } catch { return { ok: false, code:'network', message: 'Қосылым үзілді' }; }
 }
 async updateProfile(accountId:string,displayName:string){return this.authenticated('profile',{account_id:accountId,display_name:displayName});}
 async logout(){const result=await this.authenticated('logout',{});if(result.ok)await this.accountChanged();return result;}
}
