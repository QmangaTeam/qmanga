import {expect,test} from 'bun:test';
import {ReadingSession,type Progress,type ProgressStorage,type ReadingGateway,type LocalProgress} from './progress';
import type {Manifest} from './reader';
test('account progress is scoped, never auto imports guests and stops on a conflict',async()=>{
 const cloud={version:2,pageId:'cloud-page',releaseId:'release',fraction:.3} as Progress;let reads:string[]=[],writes:LocalProgress[]=[],sent=0,stopped:()=>void=()=>{};
 const storage:ProgressStorage={read:async(scope)=>{reads.push(scope);return null;},write:async row=>{writes.push(structuredClone(row));},list:async()=>[],recent:async()=>[],clearAccounts:async()=>{},identity:()=>({id:'event',deviceId:'device',sequence:1}),watchAccountChange:callback=>{stopped=callback;return ()=>{};}};
 const gateway={account:async()=> 'account-a',progress:async()=>cloud,advance:async()=>{sent++;return {outcome:'conflict',progress:{...cloud,version:3}};}} as unknown as ReadingGateway;
 const session=new ReadingSession(gateway,storage,{later:()=>()=>{}},()=>{}),manifest={titleId:'title',titleName:'Атау',titleSlug:'title',releaseId:'release',revisionId:'revision',pages:[{id:'page'}]} as Manifest;
 await session.start(manifest);await session.record(manifest,{pageId:'page',fraction:.6});await session.flush();await session.flush();
 expect(reads).toEqual(['account-a']);expect(writes[0].scope).toBe('account-a');expect(sent).toBe(1);expect(session.row?.conflict?.version).toBe(3);
 stopped();await session.record(manifest,{pageId:'page',fraction:.9});expect(session.row).toBeNull();
});
