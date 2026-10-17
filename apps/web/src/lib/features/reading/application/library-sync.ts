import type {LibraryEntry,ReadingGateway,Selection,LibraryFilter,LibraryPage} from './progress';
import {AppFailure} from '$lib/shared/domain/failure';
export interface LibraryLocal {changeId:string;account:string;selection:Selection;entry?:LibraryEntry;pending:boolean;conflict:Selection|null;error?:string}
export interface LibraryStorage {replace(row:LibraryLocal,expected:string):Promise<boolean>;read(account:string,title:string):Promise<LibraryLocal|null>;write(row:LibraryLocal):Promise<void>;list(account:string):Promise<LibraryLocal[]>}
export class LibrarySync {
 private chain:Promise<void>=Promise.resolve();
 private listeners=new Set<()=>void>();
 constructor(private gateway:ReadingGateway,private storage:LibraryStorage,private identify:()=>string){}
 subscribe(callback:()=>void){this.listeners.add(callback);return ()=>{this.listeners.delete(callback);};}
 private changed(){for(const callback of this.listeners)callback();}
 private serialized(action:()=>Promise<void>){const next=this.chain.then(action);this.chain=next.catch(()=>{});return next;}
 state(account:string,title:string){return this.storage.read(account,title);}
 rows(account:string){return this.storage.list(account);}
 remember(account:string,selection:Selection){return this.serialized(async()=>{const existing=await this.storage.read(account,selection.titleId);if(existing?.pending||(existing?.selection.version||0)>(selection.version||0))return;await this.storage.replace({changeId:this.identify(),account,selection:{...selection},entry:'name'in selection?selection as LibraryEntry:existing?.entry,pending:false,conflict:null},existing?.changeId||'');});}
 save(account:string,input:Selection){const selection:Selection={titleId:input.titleId,status:input.status,favorite:input.favorite,preferredTeamId:input.preferredTeamId,notify:input.notify,version:input.version??0,removed:input.removed===true};return this.serialized(async()=>{const old=await this.state(account,selection.titleId);await this.storage.write({changeId:this.identify(),account,selection,entry:old?.entry,pending:true,conflict:old?.conflict||null});this.changed();await this.send(account,selection.titleId);});}
 flush(account:string){return this.serialized(async()=>{for(const row of await this.rows(account))if(row.pending&&!row.conflict)await this.send(account,row.selection.titleId);});}
 private async send(account:string,title:string){
  const row=await this.state(account,title);if(!row?.pending||row.conflict)return;
  try{
   await this.gateway.save(account,row.selection);
   const cloud=await this.gateway.selection(account,title);
   if(!cloud)throw new AppFailure('network');
   row.selection={...cloud};row.entry=cloud;row.pending=false;row.error='';
  }catch(failure){
   const code=failure instanceof AppFailure?failure.code:'network';
   if(code==='conflict'){const cloud=await this.gateway.selection(account,title);if(cloud)row.conflict={...cloud};else row.error='conflict';}
   else row.error=code;
  }
  const expected=row.changeId;row.changeId=this.identify();await this.storage.replace(row,expected);this.changed();
 }
 resolve(account:string,title:string,useLocal:boolean){return this.serialized(async()=>{const row=await this.state(account,title);if(!row)return;
  if(useLocal&&row.conflict){row.selection.version=row.conflict.version;row.pending=true;}else{const cloud=await this.gateway.selection(account,title);if(cloud){row.selection={...cloud};row.entry=cloud;}row.pending=false;}
  row.conflict=null;row.error='';const expected=row.changeId;row.changeId=this.identify();if(!await this.storage.replace(row,expected))return;this.changed();if(useLocal)await this.send(account,title);
 });}
}
export class CachedReadingGateway implements ReadingGateway {
 constructor(private remote:ReadingGateway,private sync:LibrarySync){}
 account(){return this.remote.account();}
 recent(account:string,limit=4){return this.remote.recent(account,limit);}
 references(releases:string[]){return this.remote.references(releases);}
 progress(account:string,title:string){return this.remote.progress(account,title);}
 advance(...args:Parameters<ReadingGateway['advance']>){return this.remote.advance(...args);}
 async selection(account:string,title:string){let cloud:LibraryEntry|null=null;try{cloud=await this.remote.selection(account,title);if(cloud)await this.sync.remember(account,cloud);}catch(failure){if(!(failure instanceof AppFailure)||!['network','unavailable'].includes(failure.code))throw failure;const row=await this.sync.state(account,title);if(!row?.entry)throw failure;cloud=row.entry;}const row=await this.sync.state(account,title);return row?.pending?{...(cloud||row.entry||{name:'',slug:'',unseen:0,progress:{releaseId:''}}),...row.selection} as LibraryEntry:cloud;}
 async list(account:string,after='',filter:LibraryFilter={}):Promise<LibraryPage>{
  let result:LibraryPage;
  const matches=(item:LibraryEntry,query=false)=>!item.removed&&(!filter.status||item.status===filter.status)&&(!filter.favorite||item.favorite)&&(!query||!filter.query||(item.name+' '+item.slug).toLocaleLowerCase('kk').includes(filter.query.trim().toLocaleLowerCase('kk')));
  try{result=await this.remote.list(account,after,filter);for(const item of result.items)await this.sync.remember(account,item);}
  catch(failure){if(!(failure instanceof AppFailure)||!['network','unavailable'].includes(failure.code))throw failure;const cached=(await this.sync.rows(account)).filter(row=>row.entry).map(row=>({...row.entry!,...row.selection})).filter(item=>matches(item,true)&&(!after||item.titleId<after)).sort((left,right)=>right.titleId.localeCompare(left.titleId));const items=cached.slice(0,100);return {items,hasMore:cached.length>100,nextCursor:cached.length>100?items.at(-1)!.titleId:'',partial:true};}
  const pending=(await this.sync.rows(account)).filter(row=>row.pending),byTitle=new Map(pending.map(row=>[row.selection.titleId,row]));
  const items=result.items.map(item=>({...item,...byTitle.get(item.titleId)?.selection})).filter(item=>matches(item));
  if(!after)for(const row of pending)if(row.entry&&!items.some(item=>item.titleId===row.selection.titleId)){const item={...row.entry,...row.selection};if(matches(item,true))items.push(item);}
  return {...result,items,total:pending.length?undefined:result.total,partial:pending.length>0};
 }
 save(account:string,input:Selection){return this.sync.save(account,{...input,removed:input.removed===true});}
 async remove(account:string,title:string){const row=await this.selection(account,title);if(row)await this.sync.save(account,{...row,removed:true});}
}
