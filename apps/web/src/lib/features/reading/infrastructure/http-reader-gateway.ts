import type { Fragment, Manifest, ReaderGateway, Tier } from '../application/reader';
import { ApiClient } from '$lib/shared/infrastructure/api-client';
function record(value: unknown): Record<string,unknown> { if(!value||typeof value!=='object')throw new Error('Тарау деректері жарамсыз');return value as Record<string,unknown>; }
function text(value: unknown): string { if(typeof value!=='string')throw new Error('Тарау деректері жарамсыз');return value; }
function number(value: unknown): number { if(typeof value!=='number'||!Number.isFinite(value))throw new Error('Тарау деректері жарамсыз');return value; }
export function fragmentFrom(value: unknown): Fragment {
 const item=record(value),url=text(item.url),width=number(item.width),height=number(item.height),bytes=number(item.bytes),checksum=text(item.checksum),mime=text(item.mime);
 if(!/^\/api\/v1\/media\/[a-zA-Z0-9-]+$/.test(url)||width<=0||height<=0||bytes<=0||!/^image\/(webp|png|jpeg)$/.test(mime)||!/^[a-f0-9]{64}$/.test(checksum))throw new Error('Сурет деректері жарамсыз');
 return {id:text(item.id),url,width,height,bytes,checksum,mime,offset:number(item.offset)};
}
export function manifestFrom(value: unknown): Manifest {
 const item=record(value),team=record(item.team);
 if(!Array.isArray(item.pages)||!item.pages.length)throw new Error('Тарауда дайын беттер жоқ');
 const pages=item.pages.map(value=>{const page=record(value),variants=record(page.variants),mapped={} as Record<Tier,Fragment[]>;for(const tier of ['eco','standard','high'] as Tier[]){if(!Array.isArray(variants[tier])||!variants[tier].length)throw new Error('Сурет сапасы дайын емес');mapped[tier]=(variants[tier] as unknown[]).map(fragmentFrom);}return {id:text(page.id),position:number(page.position),width:number(page.width),height:number(page.height),variants:mapped};});
 return {source:typeof item.source_note==='string'?item.source_note:'',credits:typeof item.credits==='string'?item.credits:'',releaseId:text(item.release_id),revisionId:text(item.revision_id),titleId:text(item.title_id),titleSlug:text(item.title_slug),titleName:text(item.title_name),chapterLabel:text(item.chapter_label),chapterOrder:number(item.chapter_order),projectId:text(item.project_id),kind:text(item.kind),direction:item.direction==='rtl'?'rtl':'ltr',team:{id:text(team.id),name:text(team.name)},previousId:text(item.previous_release_id),nextId:text(item.next_release_id),pages};
}
export class HttpReaderGateway implements ReaderGateway {
 constructor(private api: ApiClient) {}
 async manifest(releaseId: string,preview?: string) { return manifestFrom(await this.api.call(preview?`/api/v1/studio/revisions/${encodeURIComponent(preview)}/manifest`:`/api/v1/releases/${encodeURIComponent(releaseId)}/manifest`)); }
 async original(releaseId: string,pageId: string) { return fragmentFrom(await this.api.call(`/api/v1/releases/${encodeURIComponent(releaseId)}/pages/${encodeURIComponent(pageId)}/original`)); }
}
