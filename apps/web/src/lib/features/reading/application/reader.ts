import type { QualityPreference, Position } from '../domain/policies';
export type Tier = 'eco' | 'standard' | 'high';
export interface Fragment { id: string; url: string; width: number; height: number; bytes: number; offset: number; mime: string; checksum: string }
export interface SourcePage { id: string; position: number; width: number; height: number; variants: Record<Tier, Fragment[]> }
export interface Manifest {preparationMethod?:string;humanReviewed?:boolean;humanReviewer?:string;humanReviewedAt?:string; source?:string;credits?:string;coverURL?:string; releaseId: string; revisionId: string; titleId: string; titleSlug: string; titleName: string; chapterLabel: string; chapterOrder: number; projectId: string; team: {id: string; name: string}; kind: string; direction: 'ltr' | 'rtl'; previousId: string; nextId: string; pages: SourcePage[] }
export interface ReaderGateway { manifest(releaseId: string, preview?: string): Promise<Manifest>; original(releaseId: string, pageId: string): Promise<Fragment> }
export interface ImageReference { url: string }
export interface ImageTask { done: Promise<ImageReference>; cancel(): void }
export interface ImageLoader { load(fragment: Fragment,priority?:number): ImageTask; cached?(fragment:Fragment):ImageReference|undefined; release(image: ImageReference): void }
export interface Timer { later(callback: () => void, milliseconds: number): () => void }
export class MediaFailure extends Error { constructor(message: string, readonly terminal = false, readonly retryAfter = 0, readonly code='media_unavailable') { super(message); } }
export interface ReaderPreferences { quality: Exclude<QualityPreference,'original'>; mode: 'pages' | 'strip'; theme: 'dark' | 'light' | 'system'; fit: 'width' | 'height'; maxWidth: number; brightness: number; numbers: boolean; direction: 'auto' | 'ltr' | 'rtl' }
export interface Viewport { width: number; height: number; top: number }
export interface ReaderPlatform {
 libraryPromptDismissed?(account:string,title:string):boolean;
 dismissLibraryPrompt?(account:string,title:string):void;
 openRelease(releaseId:string):void;
 prefetchRelease?(releaseId:string):Promise<Manifest|null>;
 preferences(kind: string): ReaderPreferences;
 savePreferences(value: ReaderPreferences): void;
 connection(): {saveData: boolean; slowConnection: boolean; lowMemory: boolean};
 observe(stage: unknown, callback: (viewport: Viewport) => void): () => void;
 scroll(stage: unknown, top: number): void;
 panel(dialog: unknown, close: () => void): () => void;
 fullscreen(stage: unknown): Promise<void>;
 visibility(callback: (hidden: boolean) => void): () => void;
 loadPosition(releaseId: string): Position | null;
 savePosition(releaseId: string,position: Position): void;
 keyboard(callback: (action:'right'|'left'|'menu')=>void): () => void;
}
export interface LayoutItem extends Fragment { pageId: string; pageIndex: number; top: number; ratio: number }
export function layout(manifest: Manifest, tier: Tier, mode: 'pages' | 'strip', pageIndex: number): {items: LayoutItem[]; total: number} {
 const items: LayoutItem[] = []; let total = 0;
 manifest.pages.forEach((page,index) => { if (mode === 'pages' && index !== pageIndex) return; for (const fragment of page.variants[tier]) { const ratio = fragment.height / fragment.width; items.push({...fragment,pageId:page.id,pageIndex:index,top:total,ratio}); total += ratio; } });
 return {items,total};
}
export function visible(layoutItems: LayoutItem[], viewport: Viewport): {start: number; end: number; position: Position; pageIndex: number} {
 const units = viewport.top / Math.max(1,viewport.width); let start = 0;
 while (start < layoutItems.length-1 && layoutItems[start].top + layoutItems[start].ratio <= units + 1 / Math.max(1,viewport.width)) start++;
 let end = start; const bottom = units + viewport.height / Math.max(1,viewport.width);
 while (end < layoutItems.length-1 && layoutItems[end].top + layoutItems[end].ratio < bottom) end++;
 const current = layoutItems[start]; if (!current) return {start:0,end:0,position:{pageId:'',fraction:0},pageIndex:0};
 const pageItems = layoutItems.filter(item => item.pageId === current.pageId), pageTop = pageItems[0].top, pageHeight = pageItems.reduce((sum,item) => sum+item.ratio,0);
 return {start,end,position:{pageId:current.pageId,fraction:Math.max(0,Math.min(1,(units-pageTop)/pageHeight))},pageIndex:current.pageIndex};
}
export function anchorOffset(items: LayoutItem[], position: Position, width: number): number { const page = items.filter(item => item.pageId === position.pageId); if (!page.length) return 0; return width * (page[0].top + page.reduce((sum,item) => sum+item.ratio,0) * position.fraction); }
