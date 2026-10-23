import {test,expect} from 'bun:test';
import {LibraryProgressReader} from './library-progress';
import type {LibraryEntry} from './progress';
import type {Manifest} from './reader';
const entry={titleId:'title',progress:{releaseId:'release',revisionId:'revision',pageId:'p2',fraction:.5}} as LibraryEntry;
const manifest={titleId:'title',releaseId:'release',revisionId:'revision',chapterLabel:'7',pages:[{id:'p1'},{id:'p2'},{id:'p3'}]} as Manifest;
test('library progress uses saved page and matching revision, not chapter order',async()=>{
 const reader=new LibraryProgressReader({manifest:async()=>manifest});
 expect(await reader.read(entry)).toEqual({page:2,pages:3,chapter:'7',fraction:.5});
 expect(await reader.read({...entry,progress:{...entry.progress,revisionId:'old'}})).toBeNull();
 expect(await reader.read({...entry,progress:{...entry.progress,pageId:'missing'}})).toBeNull();
 expect(await reader.read({...entry,progress:{...entry.progress,releaseId:''}})).toBeNull();
});
test('library manifest metadata is deduplicated and at most four requests run together',async()=>{
 let active=0,peak=0,calls=0;const releases:(()=>void)[]=[];
 const reader=new LibraryProgressReader({manifest:async(id)=>{calls++;active++;peak=Math.max(peak,active);await new Promise<void>(resolve=>releases.push(resolve));active--;return {...manifest,releaseId:id};}});
 const tasks=Array.from({length:8},(_,i)=>reader.read({...entry,progress:{...entry.progress,releaseId:String(i)}}));
 const duplicate=reader.read({...entry,progress:{...entry.progress,releaseId:'0'}});
 expect(calls).toBe(4);
 for(let i=0;i<8;i++){releases.shift()?.();await new Promise(resolve=>setTimeout(resolve,0));}
 await Promise.all([...tasks,duplicate]);expect(calls).toBe(8);expect(peak).toBe(4);
});
test('failed public manifests do not invent percentages and can be retried',async()=>{
 let calls=0;const reader=new LibraryProgressReader({manifest:async()=>{if(++calls===1)throw Error('offline');return manifest;}});
 expect(await reader.read(entry)).toBeNull();expect((await reader.read(entry))?.fraction).toBe(.5);
});
