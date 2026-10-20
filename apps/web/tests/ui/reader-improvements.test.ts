import {test,expect} from 'bun:test';
import Fixture from './ReaderImprovements.fixture.svelte';
import {render,elements} from './markup';
test('mixed library shows compact real status outside the unchanged options dialog',async()=>{
 const html=render(Fixture).body;
 expect(await elements(html,'article > .library-status')).toHaveLength(1);
 expect(html).toContain('Прочитано');
 expect(await elements(html,'button[aria-haspopup="dialog"]')).toHaveLength(1);
});

test('unavailable catalogue retries the exact request including cursor',async()=>{
 const html=render(Fixture,{props:{screen:'catalogue'}}).body;
 expect((await elements(html,'a')).some(link=>link.href?.replaceAll('&amp;','&')==='/catalog?q=query&genre=action&sort=name&cursor=cursor')).toBe(true);
});
