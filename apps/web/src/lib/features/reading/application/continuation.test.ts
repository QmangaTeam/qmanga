import {expect,test} from 'bun:test';
import {loadContinuation} from './continuation';
import type {LibraryEntry,ReadingGateway,ProgressStorage} from './progress';

function entry(titleId:string,status='reading',releaseId='release'):LibraryEntry {
 return {titleId,name:titleId,slug:titleId,status,progress:{releaseId}} as LibraryEntry;
}
test('continuation looks past non-reading library pages and stops at the requested limit',async()=>{
 const cursors:(string|undefined)[]=[];
 const gateway={account:async()=>'account',list:async(account:string,after?:string)=>{
  expect(account).toBe('account');cursors.push(after);
  return after?{items:[entry('second'),entry('third')],hasMore:true}:{items:[entry('planned','planned')],hasMore:true};
 }} as ReadingGateway;
 expect((await loadContinuation(gateway,{} as ProgressStorage,2)).map(row=>row.titleId)).toEqual(['second','third']);
 expect(cursors).toEqual([undefined,'planned']);
});
test('continuation terminates on an empty or non-advancing page',async()=>{
 let calls=0;
 const gateway={account:async()=>'account',list:async()=>{calls++;return {items:[entry('planned','planned')],hasMore:true};}} as ReadingGateway;
 expect(await loadContinuation(gateway,{} as ProgressStorage)).toEqual([]);
 expect(calls).toBe(2);
 gateway.list=async()=>({items:[],hasMore:true});
 expect(await loadContinuation(gateway,{} as ProgressStorage)).toEqual([]);
});
test('guest continuation never requests account library and excludes missing releases',async()=>{
 const gateway={account:async()=>null,list:async()=>{throw Error('account request');}} as unknown as ReadingGateway;
 const storage={list:async(scope:string)=>{expect(scope).toBe('guest');return [entry('missing','reading',''),entry('saved')];}} as unknown as ProgressStorage;
 expect((await loadContinuation(gateway,storage)).map(row=>row.titleId)).toEqual(['saved']);
});
