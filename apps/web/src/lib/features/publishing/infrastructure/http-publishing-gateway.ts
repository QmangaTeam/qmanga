import {AppFailure} from '../../../shared/domain/failure';
import type { PublishingGateway, Revision, UploadSource, UploadTask } from '../application/publishing';
import type { ApiClient } from '../../../shared/infrastructure/api-client';
function object(value: unknown): Record<string, unknown> { if (typeof value !== 'object' || value === null) throw new AppFailure('unavailable'); return value as Record<string, unknown>; }
function text(value: unknown) { if (typeof value !== 'string') throw new AppFailure('unavailable'); return value; }
function number(value: unknown) { if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) throw new AppFailure('unavailable'); return value; }
function array<Item>(value: unknown, map: (item: unknown) => Item): Item[] { if (!Array.isArray(value)) throw new AppFailure('unavailable'); return value.map(map); }
function revision(value: unknown): Revision {
 const item = object(value);
 return {coverId:typeof item.cover_id==='string'?item.cover_id:'',humanReviewer:typeof item.human_reviewer==='string'?item.human_reviewer:'',humanReviewedAt:typeof item.human_reviewed_at==='string'?item.human_reviewed_at:'',releaseStatus:typeof item.release_status==='string'?item.release_status:'draft',preparationMethod:typeof item.preparation_method==='string'?item.preparation_method:'unspecified',humanReviewed:item.human_reviewed===true,source:typeof item.source_note==='string'?item.source_note:'',credits:typeof item.credits==='string'?item.credits:'',rightsConfirmed:item.rights_confirmed===true,confirmedBy:typeof item.confirmed_by==='string'?item.confirmed_by:'', id: text(item.id), releaseId: text(item.release_id), projectId: text(item.project_id), chapterId: text(item.chapter_id), titleId: text(item.title_id), titleName: text(item.title_name), chapterLabel: text(item.chapter_label), teamName: text(item.team_name), state: text(item.state), note: text(item.note), version: number(item.version), pages: array(item.pages, value => { const page = object(value); return { originalName:typeof page.original_name==='string'?page.original_name:'',thumbnailId:typeof page.thumbnail_id==='string'?page.thumbnail_id:'',id: text(page.id), position: number(page.position), width: number(page.width), height: number(page.height), bytes: number(page.bytes), status: text(page.status), failure: text(page.failure), uploadKey: text(page.upload_key) }; }) };
}
export class HttpPublishingGateway implements PublishingGateway {
 constructor(private api: ApiClient) {}
 private async items<Item>(path: string, map: (value: unknown) => Item) { return array(object(await this.api.call(path)).items, map); }
 async projectPage(filter:{query?:string;teamId?:string},cursor=''){const query=new URLSearchParams({q:filter.query||'',team:filter.teamId||'',cursor});const result=object(await this.api.call('/api/v1/studio/projects?'+query));return {items:array(result.items,value=>{const item=object(value);return {id:text(item.id),titleName:text(item.title_name),teamName:text(item.team_name),teamId:text(item.team_id),coverURL:typeof item.cover_url==='string'?item.cover_url:undefined,nextRevisionId:typeof item.next_revision_id==='string'?item.next_revision_id:undefined,nextRevisionState:typeof item.next_revision_state==='string'?item.next_revision_state:undefined,chapterCount:item.chapter_count===undefined?undefined:number(item.chapter_count),ratingCount:item.rating_count===undefined?undefined:number(item.rating_count),score:typeof item.score==='number'&&Number.isFinite(item.score)&&item.score>=0&&item.score<=10?item.score:undefined};}),nextCursor:text(result.next_cursor),total:number(result.total)};}
 async project(id:string){const item=object(await this.api.call('/api/v1/studio/projects/'+encodeURIComponent(id)));return {id:text(item.id),titleId:text(item.title_id),titleName:text(item.title_name),teamId:text(item.team_id),teamName:text(item.team_name),role:text(item.role),approved:item.approved===true,status:text(item.status),coverId:text(item.cover_id)};}
 async createChapter(id:string,input:{label:string;volume:string;name:string;order:number}){return text(object(await this.api.call('/api/v1/studio/projects/'+encodeURIComponent(id)+'/chapters','POST',input)).id);}
 async projects(offset=0) { return this.items('/api/v1/studio/projects?offset='+offset, value => { const item = object(value); return { id: text(item.id), titleName: text(item.title_name), teamName: text(item.team_name),teamId:text(item.team_id) }; }); }
 async chapters(id: string,offset=0) { return this.items('/api/v1/studio/projects/' + encodeURIComponent(id) + '/chapters?offset='+offset, value => { const item = object(value); return { id: text(item.id), label: text(item.label), name: text(item.name), order: number(item.order) }; }); }
 async revisions(id: string,offset=0) { return this.items('/api/v1/studio/projects/' + encodeURIComponent(id) + '/revisions?offset='+offset, revision); }
 async revision(id: string) { return revision(await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id))); }
 async create(projectId: string, chapterId: string) { return revision(await this.api.call('/api/v1/studio/projects/' + encodeURIComponent(projectId) + '/revisions', 'POST', { chapter_id: chapterId })); }
 async discard(id:string,version:number) { await this.api.call('/api/v1/studio/revisions/'+encodeURIComponent(id)+'?version='+version,'DELETE'); }
 importChapter(file:unknown,progress:(done:number,total:number)=>void){
  const controller=new AbortController();
  let abort!:()=>void;
  const stopped=new Promise<never>((_,reject)=>{abort=()=>reject(new AppFailure('upload_cancelled'));controller.signal.addEventListener('abort',abort,{once:true});});
  const importing=(async()=>{
   if(!(file instanceof File))throw new AppFailure('upload_file');
   const options={signal:controller.signal,onProgress:progress};
   const importer=/\.pdf$/i.test(file.name)?(await import('./pdf-import')).importPdf:(await import('./cbz-import')).importCbz;
   if(controller.signal.aborted)throw new AppFailure('upload_cancelled');
   const files=await importer(file,options);
   if(controller.signal.aborted)throw new AppFailure('upload_cancelled');
   const sources=await this.prepare(files);
   if(controller.signal.aborted)throw new AppFailure('upload_cancelled');
   for(let i=0;i<sources.length;i++){
    if(controller.signal.aborted)throw new AppFailure('upload_cancelled');
    const digest=await crypto.subtle.digest('SHA-256',await files[i].arrayBuffer());
    if(controller.signal.aborted)throw new AppFailure('upload_cancelled');
    sources[i].key='import:'+Array.from(new Uint8Array(digest),byte=>byte.toString(16).padStart(2,'0')).join('')+':'+sources[i].key.slice(0,32);
   }
   return sources;
  })();
  const done=Promise.race([importing,stopped]).finally(()=>controller.signal.removeEventListener('abort',abort));
  return {done,cancel:()=>controller.abort()};
 }
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
    request.open('PUT', '/api/v1/studio/revisions/' + encodeURIComponent(id) + '/pages?key=' + encodeURIComponent(source.key) + '&position=' + position + '&name=' + encodeURIComponent(source.name));
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
 async remove(id: string, pageId: string,version:number) { await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id) + '/pages/' + encodeURIComponent(pageId)+'?version='+version, 'DELETE'); }
 async attribution(id:string,version:number,source:string,credits:string,confirmed:boolean,preparationMethod='unspecified',humanReviewed=false){await this.api.call('/api/v1/studio/revisions/'+encodeURIComponent(id)+'/attribution','PUT',{version,source_note:source,credits,rights_confirmed:confirmed,preparation_method:preparationMethod,human_reviewed:humanReviewed});}
 async submit(id: string, version: number) { await this.api.call('/api/v1/studio/revisions/' + encodeURIComponent(id) + '/submit', 'POST', { version }); }
 async pending(offset=0,team='',kind='') { return this.items('/api/v1/moderation/revisions?offset='+offset+'&team='+encodeURIComponent(team)+'&kind='+encodeURIComponent(kind), revision); }
 async review(id: string, decision: 'publish' | 'reject', reason: string) { await this.api.call('/api/v1/moderation/revisions/' + encodeURIComponent(id) + '/review', 'POST', { decision, reason }); }
 async selectCover(projectId:string,pageId:string){await this.api.call('/api/v1/studio/projects/'+encodeURIComponent(projectId)+'/cover','PUT',{page_id:pageId});}
 async uploadCover(projectId:string,file:unknown):Promise<string>{
  if(!(file instanceof File)||file.size===0||!['image/jpeg','image/png','image/webp'].includes(file.type))throw new AppFailure('upload_file');
  if(file.size>25*1024*1024)throw new AppFailure('upload_size');
  const token=await this.api.csrfToken();
  let response:Response;
  try{response=await fetch('/api/v1/studio/projects/'+encodeURIComponent(projectId)+'/cover-file',{method:'PUT',credentials:'same-origin',headers:{'X-CSRF-Token':token,'Content-Type':file.type},body:file,signal:AbortSignal.timeout(600000)});}
  catch{throw new AppFailure('network');}
  let result:Record<string,unknown>;
  try{result=object(await response.json());}catch{throw new AppFailure('unavailable');}
  if(!response.ok)throw new AppFailure(text(object(result.error).code));
  return text(result.cover_id);
 }
 async retryPage(revisionId:string,pageId:string){await this.api.call('/api/v1/studio/revisions/'+encodeURIComponent(revisionId)+'/pages/'+encodeURIComponent(pageId)+'/retry','POST',{});}
 async restore(id:string,reason:string){await this.api.call('/api/v1/moderation/releases/'+encodeURIComponent(id)+'/restore','POST',{reason});}
 async withdraw(id: string, reason: string) { await this.api.call('/api/v1/moderation/releases/' + encodeURIComponent(id) + '/withdraw', 'POST', { reason }); }
}
