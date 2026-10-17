import {AppFailure} from '../../../shared/domain/failure';
import type { PublishingGateway, Revision, UploadSource, UploadTask } from '../application/publishing';
import type { ApiClient } from '../../../shared/infrastructure/api-client';
function object(value: unknown): Record<string, unknown> { if (typeof value !== 'object' || value === null) throw new AppFailure('unavailable'); return value as Record<string, unknown>; }
function text(value: unknown) { if (typeof value !== 'string') throw new AppFailure('unavailable'); return value; }
function number(value: unknown) { if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) throw new AppFailure('unavailable'); return value; }
function array<Item>(value: unknown, map: (item: unknown) => Item): Item[] { if (!Array.isArray(value)) throw new AppFailure('unavailable'); return value.map(map); }
function revision(value: unknown): Revision {
 const item = object(value);
 return {source:typeof item.source_note==='string'?item.source_note:'',credits:typeof item.credits==='string'?item.credits:'',rightsConfirmed:item.rights_confirmed===true,confirmedBy:typeof item.confirmed_by==='string'?item.confirmed_by:'', id: text(item.id), releaseId: text(item.release_id), projectId: text(item.project_id), chapterId: text(item.chapter_id), titleId: text(item.title_id), titleName: text(item.title_name), chapterLabel: text(item.chapter_label), teamName: text(item.team_name), state: text(item.state), note: text(item.note), version: number(item.version), pages: array(item.pages, value => { const page = object(value); return { id: text(page.id), position: number(page.position), width: number(page.width), height: number(page.height), bytes: number(page.bytes), status: text(page.status), failure: text(page.failure), uploadKey: text(page.upload_key) }; }) };
}
export class HttpPublishingGateway implements PublishingGateway {
 constructor(private api: ApiClient) {}
 private async items<Item>(path: string, map: (value: unknown) => Item) { return array(object(await this.api.call(path)).items, map); }
 async projects() { return this.items('/api/v1/studio/projects', value => { const item = object(value); return { id: text(item.id), titleName: text(item.title_name), teamName: text(item.team_name) }; }); }
 async chapters(id: string) { return this.items('/api/v1/studio/projects/' + encodeURIComponent(id) + '/chapters', value => { const item = object(value); return { id: text(item.id), label: text(item.label), name: text(item.name), order: number(item.order) }; }); }
 async revisions(id: string) { return this.items('/api/v1/studio/projects/' + encodeURIComponent(id) + '/revisions', revision); }
 async revision(id: string) { return revision(await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id))); }
 async create(projectId: string, chapterId: string) { return revision(await this.api.call('/api/v1/studio/projects/' + encodeURIComponent(projectId) + '/revisions', 'POST', { chapter_id: chapterId })); }
 async prepare(files: unknown[]): Promise<UploadSource[]> {
  return Promise.all(files.map(async source => {
   if (!(source instanceof File) || source.size <= 0 || source.size > 25 * 1024 * 1024) throw new AppFailure('upload_size');
   const metadata = new TextEncoder().encode(source.name + ':' + source.size + ':' + source.lastModified);
   const digest = await crypto.subtle.digest('SHA-256', metadata);
   const key = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, '0')).join('');
   return { name: source.name, size: source.size, key, content: source };
  }));
 }
 upload(id: string, source: UploadSource, position: number, progress: (fraction: number) => void): UploadTask {
  const request = new XMLHttpRequest();
  let cancelled = false;
  const done = (async () => {
   const token = await this.api.csrfToken();
   if (cancelled) throw new AppFailure('upload_cancelled');
   return new Promise<void>((resolve, reject) => {
    request.open('PUT', '/api/v1/studio/revisions/' + encodeURIComponent(id) + '/pages?key=' + encodeURIComponent(source.key) + '&position=' + position);
    request.withCredentials = true; request.timeout = 600000;
    request.setRequestHeader('X-CSRF-Token', token); request.setRequestHeader('Content-Type', 'application/octet-stream');
    request.upload.onprogress = event => { if (event.lengthComputable) progress(event.loaded / event.total); };
    request.onerror = request.ontimeout = () => reject(new AppFailure('network'));
    request.onabort = () => reject(new AppFailure('upload_cancelled'));
    request.onload = () => {
     if (request.status === 200 || request.status === 201) { resolve(); return; }
     try { const result = object(JSON.parse(request.responseText)); reject(new AppFailure(text(object(result.error).code))); } catch { reject(new AppFailure('unavailable')); }
    };
    if (!(source.content instanceof Blob)) { reject(new AppFailure('upload_file')); return; }
    request.send(source.content);
   });
  })();
  return { done, cancel: () => { cancelled = true; request.abort(); } };
 }
 async reorder(id: string, version: number, pageIds: string[]) { await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id) + '/order', 'PUT', { version, page_ids: pageIds }); }
 async remove(id: string, pageId: string) { await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id) + '/pages/' + encodeURIComponent(pageId), 'DELETE'); }
 async attribution(id:string,version:number,source:string,credits:string,confirmed:boolean){await this.api.call('/api/v1/studio/revisions/'+encodeURIComponent(id)+'/attribution','PUT',{version,source_note:source,credits,rights_confirmed:confirmed});}
 async submit(id: string, version: number) { await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id) + '/submit', 'POST', { version }); }
 async pending() { return this.items('/api/v1/moderation/revisions', revision); }
 async review(id: string, decision: 'publish' | 'reject', reason: string) { await this.api.call('/api/v1/moderation/revisions/' + encodeURIComponent(id) + '/review', 'POST', { decision, reason }); }
 async withdraw(id: string, reason: string) { await this.api.call('/api/v1/moderation/releases/' + encodeURIComponent(id) + '/withdraw', 'POST', { reason }); }
}
