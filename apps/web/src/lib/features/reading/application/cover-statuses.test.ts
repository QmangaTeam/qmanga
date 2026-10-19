import {test,expect} from 'bun:test';
import {CoverStatuses} from './cover-statuses';
import type {ReadingGateway} from './progress';
test('cover statuses clear at account change and ignore an old account response',async()=>{
 let account:string|null='first',resolve!:(value:Record<string,string>)=>void;
 const gateway={account:async()=>account,statuses:async()=>new Promise<Record<string,string>>(done=>resolve=done)} as ReadingGateway;
 const statuses=new CoverStatuses(gateway,()=>{});
 const first=statuses.load(['one','one']);await Promise.resolve();resolve({one:'reading'});await first;expect(statuses.items).toEqual({one:'reading'});
 const pending=statuses.load(['one']);await Promise.resolve();account=null;await statuses.load(['one']);expect(statuses.items).toEqual({});resolve({one:'dropped'});await pending;expect(statuses.items).toEqual({});
});
test('cover statuses request only unique visible ids in bounded batches and recover without inventing values',async()=>{
 const calls:string[][]=[];
 const gateway={account:async()=>'owner',statuses:async(_account:string,ids:string[])=>{calls.push(ids);return {[ids[0]]:'paused'};}} as ReadingGateway;
 const statuses=new CoverStatuses(gateway,()=>{});
 await statuses.load(Array.from({length:80},(_,i)=>'title-'+i));expect(calls.map(ids=>ids.length)).toEqual([60,20]);expect(statuses.items).toEqual({'title-0':'paused','title-60':'paused'});
 gateway.statuses=async()=>{throw Error('offline');};await statuses.load(['one']);expect(statuses.items).toEqual({});
 statuses.close();await statuses.load(['one']);expect(statuses.items).toEqual({});
});
