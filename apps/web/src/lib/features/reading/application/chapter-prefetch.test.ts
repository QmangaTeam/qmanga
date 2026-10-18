import {expect,test} from 'bun:test';
import {prefetchChapter} from './chapter-prefetch';
import type {Fragment,Manifest} from './reader';

const fragment=(id:string,bytes=1024):Fragment=>({id,url:'/api/v1/media/'+id,width:100,height:100,bytes,checksum:'a'.repeat(64),mime:'image/webp',offset:0});
const manifest={pages:[1,2,3].map(index=>({id:String(index),position:index,width:100,height:100,variants:{eco:[fragment('eco-'+index)],standard:[fragment('standard-'+index)],high:[fragment('high-'+index)]}}))} as Manifest;

test('prefetches chapter pages in order and releases references into cache',async()=>{
 const loaded:string[]=[],released:string[]=[];
 await prefetchChapter(manifest,'eco',{load:item=>{loaded.push(item.id);return {done:Promise.resolve({url:item.url}),cancel(){}};},release:image=>released.push(image.url)},new AbortController().signal,5000);
 expect(loaded).toEqual(['eco-1','eco-2','eco-3']);
 expect(released).toHaveLength(3);
});

test('stops at the budget and skips optional work on constrained connections',async()=>{
 const loaded:string[]=[];
 const loader={load:(item:Fragment)=>{loaded.push(item.id);return {done:Promise.resolve({url:item.url}),cancel(){}};},release:()=>{}};
 await prefetchChapter(manifest,'eco',loader,new AbortController().signal,1500);
 expect(loaded).toEqual(['eco-1']);
});
