import {test,expect} from 'bun:test';
import {searchTaxonomy} from './taxonomy-search';
test('finds Kazakh Russian and slug without changing IDs',()=>{const items=[{id:'one',name:'Қорқыныш',nameRu:'Ужасы',slug:'horror',categorySlug:'genre'},{id:'two',name:'Мектеп',nameRu:'Школа',slug:'school'}];expect(searchTaxonomy(items,'ужас').map(x=>x.id)).toEqual(['one']);expect(searchTaxonomy(items,'ҚОРҚ').map(x=>x.id)).toEqual(['one']);expect(searchTaxonomy(items,'school').map(x=>x.id)).toEqual(['two']);expect(searchTaxonomy(items,'')).toEqual(items);expect(items[0].id).toBe('one');});
