import {test,expect} from 'bun:test';
import {receiveReports,resolveReport,type ReportQueue} from './report-queue';
import type {ModerationReport} from './community';

test('resolving every loaded row retains the fetched cursor and loads the next page',()=>{
 let queue:ReportQueue={items:[],next:'',hasMore:false};
 queue=receiveReports(queue,{items:[{id:'first'},{id:'last'}] as ModerationReport[],next:'last',hasMore:true},false);
 queue=resolveReport(resolveReport(queue,'last'),'first');
 expect(queue.items).toEqual([]);
 expect(queue.next).toBe('last');
 expect(queue.hasMore).toBe(true);
 queue=receiveReports(queue,{items:[{id:'next'}] as ModerationReport[],next:'',hasMore:false},true);
 expect(queue.items.map(item=>item.id)).toEqual(['next']);
 expect(queue.hasMore).toBe(false);
});
