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
