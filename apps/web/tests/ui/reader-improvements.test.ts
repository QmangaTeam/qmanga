import {test,expect} from 'bun:test';
import Fixture from './ReaderImprovements.fixture.svelte';
import {render,elements} from './markup';
test('mixed library shows compact real status outside the unchanged options dialog',async()=>{
 const html=render(Fixture).body;
 expect(await elements(html,'article > .library-status')).toHaveLength(1);
 expect(html).toContain('Прочитано');
 expect(await elements(html,'button[aria-haspopup="dialog"]')).toHaveLength(1);
});
