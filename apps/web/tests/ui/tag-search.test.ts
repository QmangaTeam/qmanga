import {test,expect} from 'bun:test';
import TagSearch from './TagSearch.fixture.svelte';
import {render,elements} from './markup';

test('tag search exposes a labelled combobox and submits the selected tags',async()=>{
 const html=render(TagSearch,{props:{selected:['horror']}}).body;
 const input=(await elements(html,'input[type=search]'))[0];
 expect(input.role).toBe('combobox');
 expect(input['aria-expanded']).toBe('false');
 expect(html).toContain('Тегтер');
 const hidden=await elements(html,'input[type=hidden]');
 expect(hidden).toHaveLength(1);
 expect(hidden[0].name).toBe('genre');
 expect(hidden[0].value).toBe('horror');
 expect(html).toContain('Қорқыныш');
});

test('tag names follow the interface language',()=>{
 expect(render(TagSearch,{props:{selected:['long-hair']}}).body).toContain('Ұзын шаш');
 const russian=render(TagSearch,{props:{language:'ru',selected:['long-hair']}}).body;
 expect(russian).toContain('Длинные волосы');
 expect(russian).toContain('Теги');
 expect(russian).not.toContain('Ұзын шаш');
});
