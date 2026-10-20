import {AppFailure} from '../../../shared/domain/failure';
import { naturalOrder } from '../domain/revision';
export interface StudioPage { originalName?:string; thumbnailId?:string; id: string; position: number; width: number; height: number; bytes: number; status: string; failure: string; uploadKey: string }
export interface Revision { coverId?:string;humanReviewer?:string;humanReviewedAt?:string; releaseStatus?:string;preparationMethod?:string;humanReviewed?:boolean; source?:string;credits?:string;rightsConfirmed?:boolean;confirmedBy?:string; id: string; releaseId: string; projectId: string; chapterId: string; titleId: string; titleName: string; chapterLabel: string; teamName: string; state: string; note: string; version: number; pages: StudioPage[] }
export interface ChapterChoice { id: string; label: string; name: string; order: number }
export interface UploadSource { name: string; size: number; key: string; content: unknown }
export interface UploadTask { done: Promise<void>; cancel(): void }
export interface StudioProject {id:string;titleId:string;titleName:string;teamId:string;teamName:string;role:string;approved:boolean;status:string;coverId:string}
export interface ProjectListItem {nextRevisionId?:string;nextRevisionState?:string;id:string;titleName:string;teamName:string;teamId?:string;coverURL?:string;chapterCount?:number;ratingCount?:number;score?:number}
export interface ProjectPage {items:ProjectListItem[];nextCursor:string;total:number}
export interface ProjectFilter {query?:string;teamId?:string}
export interface ChapterImportTask {done:Promise<UploadSource[]>;cancel():void}
export interface PublishingGateway {
 importChapter?(file:unknown,progress:(done:number,total:number)=>void):ChapterImportTask;
 projectPage(filter:ProjectFilter,cursor?:string):Promise<ProjectPage>;
 project(id:string):Promise<StudioProject>;
 createChapter(projectId:string,input:{label:string;volume:string;name:string;order:number}):Promise<string>;
 projects(offset?:number): Promise<{id: string; titleName: string; teamName: string;teamId?:string}[]>;
 chapters(projectId: string,offset?:number): Promise<ChapterChoice[]>;
 revisions(projectId: string,offset?:number): Promise<Revision[]>;
 revision(id: string): Promise<Revision>;
 create(projectId: string, chapterId: string): Promise<Revision>;
 discard(id:string,version:number):Promise<void>;
 prepare(files: unknown[]): Promise<UploadSource[]>;
 upload(id: string, source: UploadSource, position: number, progress: (fraction: number) => void): UploadTask;
 reorder(id: string, version: number, pageIds: string[]): Promise<void>;
 remove(id: string, pageId: string,version:number): Promise<void>;
 attribution(id:string,version:number,source:string,credits:string,confirmed:boolean,preparationMethod?:string,humanReviewed?:boolean):Promise<void>;
 submit(id: string, version: number): Promise<void>;
 pending(offset?:number,team?:string,kind?:string): Promise<Revision[]>;
 review(id: string, decision: 'publish' | 'reject', reason: string): Promise<void>;
 selectCover(projectId:string,pageId:string):Promise<void>;
 uploadCover(projectId:string,file:unknown):Promise<string>;
 retryPage(revisionId:string,pageId:string):Promise<void>;
 restore(releaseId:string,reason:string):Promise<void>;
 withdraw(releaseId: string, reason: string): Promise<void>;
}
export class UploadQueue {
 private active: UploadTask | null = null;
 private cancelled = false;
 constructor(private gateway: PublishingGateway) {}
 cancel() { this.cancelled = true; this.active?.cancel(); }
 async run(revision: Revision, sources: UploadSource[], progress: (name: string, count: number, total: number, fraction: number) => void) {
  this.cancelled = false;
  const known = new Set(revision.pages.filter(page => page.status === 'queued' || page.status === 'ready').map(page => page.uploadKey));
  const slots = new Map(revision.pages.map(page => [page.uploadKey, page.position]));
  let position = Math.max(0, ...revision.pages.map(page => page.position)) + 1;
  let count = 0;
  for (const source of naturalOrder(sources)) {
   if (this.cancelled) throw new AppFailure('upload_cancelled');
   if (known.has(source.key)) { progress(source.name, ++count, sources.length, 1); continue; }
   const slot = slots.get(source.key) ?? position++;
   this.active = this.gateway.upload(revision.id, source, slot, fraction => progress(source.name, count, sources.length, fraction));
   await this.active.done;
   known.add(source.key); progress(source.name, ++count, sources.length, 1);
  }
  this.active = null;
 }
}

// Upload first: a failed transfer must never delete the page being corrected.
export async function replacePage(gateway:PublishingGateway,revision:Revision,pageId:string,source:UploadSource,progress:(fraction:number)=>void):Promise<Revision>{
 if(!revision.pages.some(page=>page.id===pageId))throw new AppFailure('conflict');
 // A replacement is a distinct upload even if the same file was previously added.
 const replacement={...source,key:('replace:'+pageId+':'+source.key).slice(0,128)};
 const existing=revision.pages.find(page=>page.uploadKey===replacement.key);
 await gateway.upload(revision.id,replacement,existing?.position??Math.max(0,...revision.pages.map(page=>page.position))+1,progress).done;
 const latest=await gateway.revision(revision.id);
 const uploaded=latest.pages.find(page=>page.uploadKey===replacement.key);
 if(!uploaded||!latest.pages.some(page=>page.id===pageId))throw new AppFailure('conflict');
 const order=latest.pages.filter(page=>page.id!==uploaded.id).map(page=>page.id===pageId?uploaded.id:page.id);
 await gateway.reorder(latest.id,latest.version,order);
 return gateway.revision(latest.id);
}
