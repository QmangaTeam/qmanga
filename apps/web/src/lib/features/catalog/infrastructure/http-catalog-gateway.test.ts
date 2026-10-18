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
