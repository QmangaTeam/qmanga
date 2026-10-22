import {test,expect} from 'bun:test';
import {filterWork,loadedCount,workPageContext} from '../../src/lib/features/metrics/application/work-queue';
import {ExistingWorkQueueGateway} from '../../src/lib/features/metrics/infrastructure/work-queue-gateway';
import {safeQueueReturn} from '../../src/lib/shared/presentation/queue-context';

test('work queue keeps unknown and partial counts distinct, filtering only loaded data',()=>{
 expect(loadedCount(null)).toBe('—');expect(loadedCount({items:[],next:''})).toBe('0');expect(loadedCount({items:[],next:'next'})).toBe('0+');
 const items=[{id:'one',title:'Қала',team:'Team A',description:'Глава 1',href:'/a'},{id:'two',title:'Other',team:'Team B',description:'Қала',href:'/b'}];
 expect(filterWork(items,' қала ','Team A').map(item=>item.id)).toEqual(['one']);expect(filterWork(items,'','Missing')).toEqual([]);
});
test('aggregate queue preserves original cursor and exact decision target across filtered pages',async()=>{
 const calls:unknown[]=[];
 const gateway=new ExistingWorkQueueGateway(
  {pending:async(offset:number)=>{calls.push(offset);return [{id:'revision / one',titleName:'Title',teamName:'Team',chapterLabel:'1',pages:[]}];}} as never,
  {queue:async(actor:string,cursor:string)=>{calls.push([actor,cursor]);return {items:[{id:'change1',title_name:'Title',team_name:'Team',requested_name:'New',created_at:'2026-10-22'}],has_more:true};}} as never,
  {list:async()=>({items:[{id:'accepted',status:'approved'},{id:'pending',status:'pending',titleName:'Title',teamName:'Team',note:'Note',createdAt:'2026-10-22'}],hasMore:true})} as never,
  {pending:async()=>[]} as never,
  {reports:async()=>({items:[{id:'closed',status:'resolved'},{id:'report1',status:'open',titleName:'Title',teamName:'Team',reason:'Reason'}],hasMore:true,next:'opaque-cursor'})} as never,
 );
 const publication=await gateway.list('publications','actor','100');expect(calls[0]).toBe(100);expect(publication.items[0].href).toBe('/admin/moderation?revision=revision%20%2F%20one');expect(publication.items[0].pages).toBeUndefined();
 const changes=await gateway.list('changes','actor','change0');expect(calls[1]).toEqual(['actor','change0']);expect(changes.next).toBe('change1');expect(new URL(changes.items[0].href,'https://test.invalid').searchParams.get('returnTo')).toBe('/admin?queue=changes');
 const projects=await gateway.list('projects','actor');expect(projects.items.map(item=>item.id)).toEqual(['pending']);expect(projects.next).toBe('pending');
 const reports=await gateway.list('reports','actor');expect(reports.items.map(item=>item.id)).toEqual(['report1']);expect(reports.next).toBe('opaque-cursor');
 expect(safeQueueReturn('/admin?queue=changes','fallback')).toBe('/admin?queue=changes');expect(safeQueueReturn('//evil.invalid/admin','fallback')).toBe('fallback');
});

test('selected aggregate rows restore destination page depth without changing target',()=>{const page=workPageContext({items:[{id:'x',title:'X',team:'',description:'',href:'/admin/reports#queue-x'},{id:'y',title:'Y',team:'',description:'',href:'/admin/moderation?revision=y'}],next:'cursor'},3);expect(page.items[0].href).toBe('/admin/reports?loaded=3#queue-x');expect(page.items[1].href).toBe('/admin/moderation?revision=y');});
