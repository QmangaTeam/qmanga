import {test,expect} from 'bun:test';
import {ApiClient} from '../../../shared/infrastructure/api-client';
import {HttpCatalogGateway} from './http-catalog-gateway';
test('release feed preserves actual release and team with bounded cursor parameters',async()=>{
 let requested='';
 const request=(async(url:RequestInfo|URL)=>{requested=String(url);return new Response(JSON.stringify({items:[{release_id:'release',title_id:'title',title_name:'Атау',title_slug:'atau',chapter_label:'1.10',team_name:'Топ',team_slug:'top',demo:false,published_at:'2026-10-17T00:00:00Z'}],next_cursor:'release'}));}) as typeof fetch;
 const gateway=new HttpCatalogGateway(new ApiClient(request));
 const result=await gateway.updates('previous','top');
 expect(requested).toContain('after=previous');
 expect(requested).toContain('team=top');
 expect(result.items[0]).toMatchObject({releaseId:'release',chapterLabel:'1.10',teamSlug:'top'});
 expect(result.nextCursor).toBe('release');
});
