import {test,expect} from 'bun:test';
import TaxonomySelection from '../../src/lib/features/catalog/presentation/TaxonomySelection.svelte';
import {render,elements} from './markup';
test('taxonomy indicates capacity and preserves unknown selected IDs for removal',async()=>{
 const html=render(TaxonomySelection,{props:{label:'Жанрлар',items:[],selected:['unknown'],limit:12,onchange:()=>{}}}).body;
 expect(html).toContain('1 / 12');expect((await elements(html,'.selected button'))[0]['aria-label']).toContain('unknown');
});
