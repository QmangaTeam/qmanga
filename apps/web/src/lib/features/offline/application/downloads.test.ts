import {test,expect} from 'bun:test';
import {Downloads,type Download,type DownloadStorage} from './downloads';
import type {Manifest,Fragment} from '../../reading/application/reader';
test('an interrupted download stays partial and resumes only verified files',async()=>{
 const fragments=[{url:'/api/v1/media/a',bytes:10},{url:'/api/v1/media/b',bytes:20}] as Fragment[],manifest={releaseId:'release',revisionId:'revision',pages:[{variants:{eco:fragments}}]} as Manifest;
 let entry:Download|undefined,fail=true;const stored=new Set<string>(),calls:string[]=[];
 const storage:DownloadStorage={list:async()=>entry?[entry]:[],write:async item=>{entry=item;},remove:async()=>{},has:async file=>stored.has(file.url),store:async file=>{calls.push(file.url);if(fail&&file.url.endsWith('b'))throw new Error('offline');stored.add(file.url);},available:async()=>null,install:async()=>{}};
 const manager=new Downloads(storage,()=>1);
 await expect(manager.save(manifest,'eco',new AbortController().signal,()=>{})).rejects.toThrow('offline');
 expect(entry?.status).toBe('error');expect(entry?.bytes).toBe(10);
 fail=false;await manager.save(manifest,'eco',new AbortController().signal,()=>{});
 expect(entry?.status).toBe('ready');expect(calls.filter(url=>url.endsWith('a'))).toHaveLength(1);
 stored.clear();expect((await manager.reconcile(entry!)).status).toBe('partial');
});

function saved(id:string,urls:string[],status:Download['status']='ready'):Download {
 const fragments=urls.map(url=>({url,bytes:10})) as Fragment[];
 return {id,manifest:{releaseId:id.split(':')[0],revisionId:'revision',pages:[{variants:{eco:fragments}}]} as Manifest,tier:'eco',status,files:urls,bytes:urls.length*10,total:urls.length*10,message:'',updatedAt:1};
}
function memory(entries:Download[],stored:Set<string>):DownloadStorage {
 return {list:async()=>entries.map(item=>({...item})),write:async item=>{const index=entries.findIndex(entry=>entry.id===item.id);if(index<0)entries.push(item);else entries[index]=item;},remove:async()=>{},has:async file=>stored.has(file.url),store:async file=>{stored.add(file.url);},available:async()=>null,install:async()=>{}};
}
test('saved list verifies every chapter and demotes evicted files before offering reading',async()=>{
 const entries=[saved('one',['a']),saved('two',['b']),saved('evicted',['c'])];
 const manager=new Downloads(memory(entries,new Set(['a','b'])),()=>2);
 const list=await manager.list();
 expect(list.map(item=>[item.id,item.status,item.bytes])).toEqual([['one','ready',10],['two','ready',10],['evicted','partial',0]]);
 expect(list[2].notice).toBe('download_partial');
});
test('offline reading requires an explicit chapter and honors the exact saved copy',async()=>{
 const entries=[saved('one:old',['a']),saved('one:new',['b']),saved('two',['c'])];entries[1].updatedAt=3;
 const manager=new Downloads(memory(entries,new Set(['a','b','c'])),()=>2);
 expect(await manager.open()).toBeNull();
 expect((await manager.open({id:'one:old'}))?.id).toBe('one:old');
 expect((await manager.open({releaseId:'two'}))?.id).toBe('two');
 expect(await manager.open({id:'absent',releaseId:'two'})).toBeNull();
});
test('an empty chapter cannot be classified as complete',async()=>{
 const entry=saved('empty',[]),manager=new Downloads(memory([entry],new Set()),()=>2);
 expect((await manager.reconcile(entry)).status).toBe('partial');
 await expect(manager.save(entry.manifest,'eco',new AbortController().signal,()=>{})).rejects.toThrow();
});
test('a file evicted while another downloads prevents a ready result',async()=>{
 const entry=saved('chapter',['a','b']),stored=new Set<string>(),storage=memory([],stored);
 storage.store=async file=>{stored.clear();stored.add(file.url);};
 const manager=new Downloads(storage,()=>2);
 await expect(manager.save(entry.manifest,'eco',new AbortController().signal,()=>{})).rejects.toThrow();
 expect((await manager.list())[0].status).not.toBe('ready');
});

 test('bulk cleanup deduplicates exact selected copies and reports failures without touching unselected entries',async()=>{
 const rows=[saved('one',['a']),saved('two',['b']),saved('three',['c'])],storage=memory(rows,new Set(['a','b','c'])),removed:string[]=[];
 storage.remove=async id=>{if(id==='two')throw Error('unavailable');removed.push(id);};
 const manager=new Downloads(storage,()=>2);
 expect(await manager.removeSelected(['one','one','two','foreign'])).toEqual({removed:['one'],failed:['two']});
 expect(removed).toEqual(['one']);expect(await manager.available()).toBeNull();
 });
