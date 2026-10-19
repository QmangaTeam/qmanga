import {test,expect} from 'bun:test';
import {RegistrationImport} from './registration-import';
import type {LocalProgress,ProgressStorage,ReadingGateway} from './progress';
const guest={scope:'guest',titleId:'title',name:'Title',slug:'title',progress:{titleId:'title',releaseId:'release',revisionId:'revision',pageId:'page',fraction:.5,version:0,furthestOrder:0,deviceId:'guest',sequence:1},pending:null,conflict:null} as LocalProgress;
function fixture(){
 const writes:unknown[]=[],saved:unknown[]=[],completed:string[]=[];
 const reading={account:async()=>'new',progress:async()=>({...guest.progress,version:0}),advance:async(actor:string,event:unknown)=>{writes.push({actor,event});return {outcome:'accepted',progress:{...guest.progress,version:1}};},selection:async()=>null,save:async(actor:string,selection:unknown)=>{saved.push({actor,selection});}} as unknown as ReadingGateway;
 const storage={list:async()=>[guest],identity:()=>({id:'event',deviceId:'device',sequence:1})} as ProgressStorage;
 const permit={allowed:async()=>true,complete:async(actor:string)=>{completed.push(actor);}};
 return {reading,storage,permit,writes,saved,completed};
}
test('only a registration receipt permits automatic import; repeats retain cloud progress',async()=>{
 const f=fixture();f.permit.allowed=async()=>false;const importer=new RegistrationImport(f.reading,f.storage,f.permit,()=>{});await importer.run();expect(f.writes).toHaveLength(0);expect(f.completed).toHaveLength(0);
 f.permit.allowed=async()=>true;importer.clear();await importer.run();expect(f.writes).toHaveLength(1);expect(f.saved).toHaveLength(1);expect(f.completed).toEqual(['new']);
 f.reading.progress=async()=>({...guest.progress,version:3});await importer.run();expect(f.writes).toHaveLength(1);
});
test('failed import is retryable, preserves guest data and cannot complete for a switched account',async()=>{
 const f=fixture();f.reading.advance=async()=>{throw Error('offline');};const importer=new RegistrationImport(f.reading,f.storage,f.permit,()=>{});await importer.run();expect(importer.failed).toBe(true);expect(f.completed).toHaveLength(0);
 let finish!:(value:ReturnType<ReadingGateway['progress']> extends Promise<infer Value>?Value:never)=>void;
 let started!:()=>void;const ready=new Promise<void>(resolve=>started=resolve);f.reading.progress=async()=>{started();return new Promise(resolve=>finish=resolve);};const pending=importer.run();await ready;importer.clear();finish({...guest.progress,version:4});await pending;expect(f.saved).toHaveLength(0);expect(f.completed).toHaveLength(0);
});

test('existing cloud progress and an explicit removed selection are never replaced',async()=>{
 const f=fixture();f.reading.progress=async()=>({...guest.progress,version:7,fraction:.9});await new RegistrationImport(f.reading,f.storage,f.permit,()=>{}).run();expect(f.writes).toHaveLength(0);
 f.saved.length=0;f.reading.selection=async()=>({removed:true}) as Awaited<ReturnType<ReadingGateway['selection']>>;await new RegistrationImport(f.reading,f.storage,f.permit,()=>{}).run();expect(f.writes).toHaveLength(0);expect(f.saved).toHaveLength(0);
});

test('offline identity lookup does not pretend an import was attempted',async()=>{const f=fixture();f.reading.account=async()=>{throw Error('offline');};const importer=new RegistrationImport(f.reading,f.storage,f.permit,()=>{});await importer.run();expect(importer.failed).toBe(false);expect(f.completed).toHaveLength(0);});

test('confirmed writes refresh reading views even if the final receipt acknowledgement is interrupted',async()=>{const f=fixture();let refreshed=0;f.permit.complete=async()=>{throw Error('lost acknowledgement');};const importer=new RegistrationImport(f.reading,f.storage,f.permit,()=>{},()=>refreshed++);await importer.run();expect(importer.failed).toBe(true);expect(refreshed).toBe(1);expect(f.writes).toHaveLength(1);});
