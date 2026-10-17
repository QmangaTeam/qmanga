function object(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null; }
export class ApiFailure extends Error {
 constructor(readonly status: number, message: string) { super(message); this.name = 'ApiFailure'; }
}
export class ApiClient {
 private csrf = '';
 constructor(private request: typeof fetch) {}
 async call(path: string, method = 'GET', body?: unknown): Promise<unknown> {
  if (!path.startsWith('/api/v1/')) throw new Error('Жарамсыз сұрау');
  if (method !== 'GET' && !this.csrf) {
   const value = await this.call('/api/v1/auth/csrf');
   if (!object(value) || typeof value.csrf_token !== 'string') throw new Error('Қайта кіріңіз');
   this.csrf = value.csrf_token;
  }
  const response = await this.request(path, {
   method, credentials: 'same-origin', signal: AbortSignal.timeout(20000),
   headers: { 'Content-Type': 'application/json', ...(method === 'GET' ? {} : { 'X-CSRF-Token': this.csrf }) },
   ...(body === undefined ? {} : { body: JSON.stringify(body) })
  });
  if (response.status === 204) return null;
  const value: unknown = await response.json();
  if (!response.ok) {
   if (response.status === 403) this.csrf = '';
   const message = object(value) && object(value.error) && typeof value.error.message === 'string' ? value.error.message : 'Сұрауды орындау мүмкін болмады';
   throw new ApiFailure(response.status, message);
  }
  return value;
 }
}
