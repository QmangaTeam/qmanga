import {test,expect} from 'bun:test';
import {ApiClient} from '../../../shared/infrastructure/api-client';
import {HttpCatalogGateway} from './http-catalog-gateway';
test('release feed preserves actual release and team with bounded cursor parameters',async()=>{
 let requested='';
 const request=(async(url:RequestInfo|URL)=>{requested=String(url);return new Response(JSON.stringify({items:[{release_id:'release',title_id:'title',title_name:'Атау',title_slug:'atau',chapter_label:'1.10',team_name:'Топ',team_slug:'top',demo:false,published_at:'2026-10-17T00:00:00Z',cover_url:'/api/v1/media/cover'}],next_cursor:'release'}));}) as typeof fetch;
 const gateway=new HttpCatalogGateway(new ApiClient(request));
 const result=await gateway.updates('previous','top');
 expect(requested).toContain('after=previous');
 expect(requested).toContain('team=top');
 expect(result.items[0]).toMatchObject({releaseId:'release',chapterLabel:'1.10',teamSlug:'top',coverURL:'/api/v1/media/cover'});
 expect(result.nextCursor).toBe('release');
});

test('release thumbnails reject storage and external URLs rather than bypassing Go',async()=>{
 for(const coverURL of ['https://storage.example/cover.webp','http://minio:9000/bucket/cover.webp']){
  const request=(async()=>new Response(JSON.stringify({items:[{release_id:'release',title_id:'title',title_name:'Атау',title_slug:'atau',chapter_label:'1',team_name:'Топ',team_slug:'top',demo:false,published_at:'2026-10-17T00:00:00Z',cover_url:coverURL}],next_cursor:null}))) as typeof fetch;
  const gateway=new HttpCatalogGateway(new ApiClient(request));
  await expect(gateway.updates()).rejects.toThrow();
 }
});

test('ranked shelves request the actual type endpoint and preserve score and cover',async()=>{
 let requested='';
 const request=(async(url:RequestInfo|URL)=>{requested=String(url);return new Response(JSON.stringify({items:[{id:'ranked',slug:'ranked',name:'Тайтл',kind:'manhwa',original_status:'ongoing',translation_status:'ongoing',cover_url:'/api/v1/media/cover',demo:false,chapter_count:1,rating_count:2,score:8.5,updated_at:'2026-10-17T00:00:00Z'}],next_cursor:null}));}) as typeof fetch;
 const gateway=new HttpCatalogGateway(new ApiClient(request));
 const result=await gateway.rankings('manhwa');
 expect(requested).toBe('/api/v1/rankings?kind=manhwa');
 expect(result.items[0]).toMatchObject({kind:'manhwa',score:8.5,ratingCount:2,coverURL:'/api/v1/media/cover'});
 expect(result.nextCursor).toBeNull();
});

test('chapter translation filter is sent together with search, sort and continuation',async()=>{
 let requested='';
 const request=(async(url:RequestInfo|URL)=>{requested=String(url);return new Response(JSON.stringify({items:[],next_cursor:null}));}) as typeof fetch;
 const gateway=new HttpCatalogGateway(new ApiClient(request));
 await gateway.chapters('example','51',{team:'exact-team',query:'1.5',sort:'desc'});
 const url=new URL(requested,'http://localhost');
 expect(url.searchParams.get('team')).toBe('exact-team');
 expect(url.searchParams.get('q')).toBe('1.5');
 expect(url.searchParams.get('sort')).toBe('desc');
 expect(url.searchParams.get('cursor')).toBe('51');
});
