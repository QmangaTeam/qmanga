import { expect, test } from 'bun:test';
import { ApiClient } from '../../../shared/infrastructure/api-client';
import { HttpCatalogGateway } from './http-catalog-gateway';

function gateway(coverURL: string) {
 const request = (async () => new Response(JSON.stringify({ items: [{ id: 'title', slug: 'dala', name: 'Дала үні', kind: 'manga', original_status: 'ongoing', translation_status: 'ongoing', cover_url: coverURL, demo: true, chapter_count: 1, rating_count: 0, score: 0, updated_at: '2026-10-17T00:00:00Z' }], next_cursor: null }), { status: 200 })) as typeof fetch;
 return new HttpCatalogGateway(new ApiClient(request));
}
test('catalogue maps transport names and permits only Go media URLs', async () => {
 const page = await gateway('/api/v1/media/asset-1').list({});
 expect(page.items[0].name).toBe('Дала үні');
 expect(page.items[0].chapterCount).toBe(1);
 expect('chapter_count' in page.items[0]).toBe(false);
 await expect(gateway('http://minio:9000/media/private.jpg').list({})).rejects.toThrow();
 await expect(gateway('//foreign.example/cover.jpg').list({})).rejects.toThrow();
});

test('public portfolio maps Telegram independently and drops non-public links',async()=>{
 for(const [link,expected] of [['https://t.me/dala_team','https://t.me/dala_team'],['https://t.me@evil.test/dala_team',''],['javascript:alert(1)','']]){
  const client=new ApiClient((async()=>Response.json({id:'team',slug:'dala',name:'Дала',description:'',telegram_url:link,members:[],projects:[]})) as typeof fetch);
  expect((await new HttpCatalogGateway(client).team('dala'))?.telegramURL).toBe(expected);
 }
});

test('catalogue carries validated age ratings and preserves the age filter including unknown',async()=>{
 for(const value of [undefined,0,12,16,18,14,'18']){
  let requested='';
  const client=new ApiClient((async(input:RequestInfo|URL)=>{
   requested=String(input);
   return Response.json({items:[{id:'title',slug:'dala',name:'Дала',kind:'manga',original_status:'ongoing',translation_status:'ongoing',cover_url:'',demo:false,chapter_count:0,rating_count:0,score:0,updated_at:'2026-10-17T00:00:00Z',...(value===undefined?{}:{age_rating:value})}],next_cursor:null});
  }) as typeof fetch);
  const gateway=new HttpCatalogGateway(client);
  if(value===14||value==='18')await expect(gateway.list({})).rejects.toThrow();
  else{
   const result=await gateway.list({ageRatings:['0']});
   expect(result.items[0].ageRating).toBe(value??0);
   expect(new URL(requested,'http://localhost').searchParams.get('age_rating')).toBe('0');
  }
 }
});

test('taxonomy maps distinct categories and ordered collections and preserves combined URL filters',async()=>{
 let requested='';
 const client=new ApiClient((async(input:RequestInfo|URL)=>{
  requested=String(input);
  if(requested.includes('/taxonomy'))return Response.json({genres:[{id:'genre',name:'Adventure',slug:'adventure'}],categories:[{id:'category',name:'Қиял әлемдері',slug:'fantasy-worlds'}],collections:[{id:'selection',name:'Selection',slug:'selection',description:'Editorial',position:2,visible:true,title_ids:['second','first']}]});
  return Response.json({items:[],next_cursor:null});
 }) as typeof fetch);
 const gateway=new HttpCatalogGateway(client),result=await gateway.taxonomy();
 expect(result.categories[0].id).toBe('category');expect(result.collections[0].titleIds).toEqual(['second','first']);expect('title_ids' in result.collections[0]).toBe(false);
 await gateway.list({category:'fantasy-worlds',collection:'selection',genres:['adventure','horror'],ageRatings:['18']},'cursor');
 const params=new URL(requested,'http://localhost').searchParams;
 const rest=new URLSearchParams(params);rest.delete('genre');
 expect(Object.fromEntries(rest)).toEqual({category:'fantasy-worlds',collection:'selection',age_rating:'18',cursor:'cursor'});
 expect(params.getAll('genre')).toEqual(['adventure','horror']);
});
