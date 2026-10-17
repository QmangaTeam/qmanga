import {test,expect} from 'bun:test';
import Product from './Product.fixture.svelte';
import {render,elements} from './markup';
test('covers retain backend URLs, native links and reserved dimensions',async()=>{
 const html=render(Product).body;
 expect((await elements(html,'img'))[0]).toMatchObject({src:'/api/v1/media/example',width:'280',height:'400',loading:'lazy',alt:''});
 expect((await elements(html,'a[aria-current]'))[0]).toMatchObject({'aria-current':'page',href:'/catalog'});
 expect(await elements(html,'a a, a button')).toHaveLength(0);
});
test('library does not invent continuation and chapter retains real release',async()=>{
 const html=render(Product).body;
 expect(await elements(html,'[data-library] a[href^="/read/"]')).toHaveLength(0);
 expect((await elements(html,'a[href^="/read/"]'))[0].href).toBe('/read/release');
 expect((await elements(render(Product,{props:{releaseId:'saved'}}).body,'[data-library] a[href^="/read/"]'))[0].href).toBe('/read/saved');
});
