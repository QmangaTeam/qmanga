import {test,expect} from 'bun:test';
import {replyCursor,mergeReplyPage} from './reply-page';
import type {Comment} from './community';
const comment=(id:string,rootId='root'):Comment=>({id,rootId,parentId:rootId,userId:'user',body:id,hidden:false,spoiler:false,createdAt:id,replies:0});
test('a focused later reply never advances the contiguous reply cursor',()=>{const root={...comment('root',''),replyCursor:'10',replies:30};const existing=[root,...Array.from({length:10},(_,i)=>comment(String(i+1))),comment('25')];expect(replyCursor(root,existing)).toBe('10');const next=mergeReplyPage(existing,root,Array.from({length:20},(_,i)=>comment(String(i+11))));expect(next.filter(item=>item.rootId==='root')).toHaveLength(30);expect(new Set(next.map(item=>item.id)).size).toBe(31);expect(next.find(item=>item.id==='root')?.replyCursor).toBe('30');});
