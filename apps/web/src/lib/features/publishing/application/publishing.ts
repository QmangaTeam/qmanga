import {AppFailure} from '../../../shared/domain/failure';
import { naturalOrder } from '../domain/revision';
export interface StudioPage { id: string; position: number; width: number; height: number; bytes: number; status: string; failure: string; uploadKey: string }
export interface Revision { id: string; releaseId: string; projectId: string; chapterId: string; titleId: string; titleName: string; chapterLabel: string; teamName: string; state: string; note: string; version: number; pages: StudioPage[] }
export interface ChapterChoice { id: string; label: string; name: string; order: number }
export interface UploadSource { name: string; size: number; key: string; content: unknown }
export interface UploadTask { done: Promise<void>; cancel(): void }
export interface PublishingGateway {
 projects(): Promise<{id: string; titleName: string; teamName: string}[]>;
 chapters(projectId: string): Promise<ChapterChoice[]>;
 revisions(projectId: string): Promise<Revision[]>;
 revision(id: string): Promise<Revision>;
 create(projectId: string, chapterId: string): Promise<Revision>;
 prepare(files: unknown[]): Promise<UploadSource[]>;
 upload(id: string, source: UploadSource, position: number, progress: (fraction: number) => void): UploadTask;
 reorder(id: string, version: number, pageIds: string[]): Promise<void>;
 remove(id: string, pageId: string): Promise<void>;
 submit(id: string, version: number): Promise<void>;
 pending(): Promise<Revision[]>;
 review(id: string, decision: 'publish' | 'reject', reason: string): Promise<void>;
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
