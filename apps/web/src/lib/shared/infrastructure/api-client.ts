function object(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null; }
import {AppFailure} from '../domain/failure';
export class ApiFailure extends AppFailure {
 constructor(readonly status: number, message: string,readonly payload?:unknown) { super(object(payload)&&object(payload.error)&&typeof payload.error.code==='string'?payload.error.code:status===0?'network':status===401?'unauthorized':'unavailable'); this.message=message; this.name = 'ApiFailure'; }
}
export class ApiClient {
 private csrf = '';
 constructor(private request: typeof fetch) {}
 async csrfToken(): Promise<string> {
  if (!this.csrf) { const value = await this.call('/api/v1/auth/csrf'); if (!object(value) || typeof value.csrf_token !== 'string') throw new Error('Қайта кіріңіз'); this.csrf = value.csrf_token; }
  return this.csrf;
 }
 async call(path: string, method = 'GET', body?: unknown,accountId=''): Promise<unknown> {
  if (!path.startsWith('/api/v1/')) throw new Error('Жарамсыз сұрау');
  if (method !== 'GET' && !this.csrf) {
   const value = await this.call('/api/v1/auth/csrf');
   if (!object(value) || typeof value.csrf_token !== 'string') throw new Error('Қайта кіріңіз');
   this.csrf = value.csrf_token;
  }
  let response:Response;
  try{response = await this.request(path, {
   method, credentials: 'same-origin', signal: AbortSignal.timeout(20000),
   headers: { 'Content-Type': 'application/json', ...(accountId?{'X-Qmanga-Account':accountId}:{}),...(method === 'GET' ? {} : { 'X-CSRF-Token': this.csrf }) },
   ...(body === undefined ? {} : { body: JSON.stringify(body) })
  });}catch{throw new ApiFailure(0,'Қосылым үзілді. Желіге қосылып, қайта көріңіз.');}
  if (response.status === 204) return null;
  let value:unknown;try{value=await response.json();}catch{throw new ApiFailure(response.status>=400?response.status:503,'Қызмет уақытша қолжетімсіз. Кейінірек қайталап көріңіз.');}
  if (!response.ok) {
   if (response.status === 403) this.csrf = '';
   const message = object(value) && object(value.error) && typeof value.error.message === 'string' ? value.error.message : 'Сұрауды орындау мүмкін болмады';
   throw new ApiFailure(response.status, message,value);
  }
  return value;
 }
}
