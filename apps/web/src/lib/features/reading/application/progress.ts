import type {Manifest,Timer} from './reader';
import type {Position} from '../domain/policies';
export interface Progress extends Position {chapterLabel?:string;teamName?:string;pagePosition?:number;updatedAt?:number;titleId:string;releaseId:string;revisionId:string;version:number;furthestOrder:number;deviceId:string;sequence:number}
export interface ProgressEvent {id:string;deviceId:string;sequence:number;baseVersion:number;titleId:string;releaseId:string;revisionId:string;pageId:string;fraction:number;completed:boolean}
export interface LocalProgress {scope:string;titleId:string;name:string;slug:string;progress:Progress;pending:ProgressEvent|null;conflict:Progress|null;updatedAt?:number;coverURL?:string;chapterLabel?:string;pagePosition?:number}
export interface LibraryEntry {version?:number;removed?:boolean;titleId:string;name:string;slug:string;status:string;favorite:boolean;preferredTeamId:string;notify:boolean;unseen:number;progress:Progress}
export interface Selection {version?:number;removed?:boolean;titleId:string;status:string;favorite:boolean;preferredTeamId:string;notify:boolean}
export interface LibraryFilter {query?:string;status?:string;favorite?:boolean}
export interface LibraryPage {items:LibraryEntry[];hasMore:boolean;nextCursor?:string;total?:number;partial?:boolean}
export interface ReadReference {teamName?:string;titleId:string;name:string;slug:string;coverURL:string;releaseId:string;revisionId:string;chapterLabel:string}
export interface RecentReading extends ReadReference {progress:Progress;updatedAt:number;pagePosition:number;availability:'available'|'updated'|'withdrawn'|'last-known'}
export interface ReadingGateway {
 account():Promise<string|null>;
 selection(accountId:string,titleId:string):Promise<LibraryEntry|null>;
 progress(accountId:string,titleId:string):Promise<Progress>;
 advance(accountId:string,event:ProgressEvent):Promise<{outcome:string;progress:Progress}>;
 list(accountId:string,after?:string,filter?:LibraryFilter):Promise<LibraryPage>;
 recent(accountId:string,limit?:number):Promise<RecentReading[]>;
 references(releases:string[]):Promise<ReadReference[]>;
 save(accountId:string,input:Selection):Promise<void>;
 remove(accountId:string,titleId:string):Promise<void>;
}
export interface ProgressStorage {read(scope:string,titleId:string):Promise<LocalProgress|null>;write(value:LocalProgress):Promise<void>;list(scope:string):Promise<LocalProgress[]>;recent(scope:string,limit:number):Promise<LocalProgress[]>;clearAccounts():Promise<void>;identity():{deviceId:string;sequence:number;id:string};watchAccountChange(callback:()=>void):()=>void}
export class ReadingSession {
 row:LocalProgress|null=null;
 notice='';private acceptedRelease='';private saving=false;private stopped=false;private cancelTimer:(()=>void)|undefined;private stopWatch:(()=>void)|undefined;
 constructor(private gateway:ReadingGateway,private storage:ProgressStorage,private timer:Timer,private changed:()=>void,private now:()=>number=Date.now){}
 async start(manifest:Manifest,guestOnly=false):Promise<Progress|null>{
  this.stopWatch=this.storage.watchAccountChange(()=>{this.close();this.row=null;this.notice="account_changed";this.changed();});
  try{const account=guestOnly?null:await this.gateway.account(),scope=account||'guest',local=await this.storage.read(scope,manifest.titleId);
   let cloud:Progress|null=null;if(account)cloud=await this.gateway.progress(account,manifest.titleId);
   if(this.stopped)return null;
   this.row=local||{scope,titleId:manifest.titleId,name:manifest.titleName,slug:manifest.titleSlug,progress:cloud||{titleId:manifest.titleId,releaseId:manifest.releaseId,revisionId:manifest.revisionId,pageId:manifest.pages[0].id,fraction:0,version:0,furthestOrder:0,deviceId:'',sequence:0},pending:null,conflict:null};
   if(cloud&&local?.pending&&cloud.version!==local.pending.baseVersion)this.row.conflict=cloud;else if(cloud&&!local?.pending)this.row.progress=cloud;
   this.notice=this.row.conflict?"progress_conflict":scope==='guest'?"progress_guest":"progress_account";this.changed();if(this.row.pending&&!this.row.conflict)void this.flush();return this.row.progress;
  }catch{this.notice="progress_auth_failed";this.changed();return null;}
 }
 async record(manifest:Manifest,position:Position,completed=false){
  if(!this.row||this.stopped||(this.acceptedRelease&&manifest.releaseId!==this.acceptedRelease))return;const identity=this.storage.identity(),baseVersion=this.row.progress.version;
  this.row.progress={...this.row.progress,...position,chapterLabel:manifest.chapterLabel,teamName:manifest.team.name,pagePosition:manifest.pages.findIndex(page=>page.id===position.pageId)+1,updatedAt:this.now(),releaseId:manifest.releaseId,revisionId:manifest.revisionId,deviceId:identity.deviceId,sequence:identity.sequence};
  this.row.updatedAt=this.now();this.row.name=manifest.titleName;this.row.slug=manifest.titleSlug;this.row.coverURL=manifest.coverURL||'';this.row.chapterLabel=manifest.chapterLabel;this.row.pagePosition=Math.max(1,manifest.pages.findIndex(page=>page.id===position.pageId)+1);
  if(this.row.scope!=='guest')this.row.pending={...identity,baseVersion,titleId:manifest.titleId,releaseId:manifest.releaseId,revisionId:manifest.revisionId,...position,completed};
  try{await this.storage.write(this.row);this.notice=this.row.conflict?"progress_conflict":this.row.scope==='guest'?"progress_local_saved":"progress_pending";this.changed();if(completed)void this.flush();else if(!this.cancelTimer)this.cancelTimer=this.timer.later(()=>{this.cancelTimer=undefined;void this.flush();},15000);}catch{this.notice="storage_unavailable";this.changed();}
 }
 async flush(){
  if(!this.row?.pending||this.row.conflict||this.saving||this.stopped)return;this.saving=true;const scope=this.row.scope,event=this.row.pending;
  try{const result=await this.gateway.advance(scope,event);if(!this.row||this.stopped||this.row.scope!==scope)return;
   if(result.outcome==='conflict'||result.outcome==='stale'){this.row.conflict=result.progress;this.notice="progress_conflict";}
   else{this.row.progress.version=result.progress.version;this.row.progress.furthestOrder=result.progress.furthestOrder;if(this.row.pending?.id===event.id)this.row.pending=null;else if(this.row.pending)this.row.pending.baseVersion=result.progress.version;this.notice=this.row.pending?"progress_pending":"progress_cloud_saved";}await this.storage.write(this.row);
  }catch{this.notice="progress_network";}finally{this.saving=false;this.changed();}
 }
 async resolve(useLocal:boolean){if(!this.row?.conflict)return;const cloud=this.row.conflict;this.row.progress.version=cloud.version;if(useLocal&&this.row.pending){const identity=this.storage.identity();this.row.pending={...this.row.pending,...identity,baseVersion:cloud.version};}else{this.row.progress=cloud;this.row.pending=null;this.acceptedRelease=cloud.releaseId;}this.row.conflict=null;await this.storage.write(this.row);this.changed();if(useLocal)await this.flush();return this.row.progress;}
 close(){this.stopped=true;this.cancelTimer?.();this.stopWatch?.();}
}
