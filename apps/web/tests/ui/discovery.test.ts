import {test,expect} from 'bun:test';
import Discovery from './Discovery.fixture.svelte';
import HomeDiscovery from './HomeDiscovery.fixture.svelte';
import {render,elements} from './markup';

test('update thumbnail opens its title while chapter opens the real release',async()=>{
 const html=render(Discovery).body;
 expect(await elements(html,'.updates a[href="/titles/example"] img')).toHaveLength(1);
 expect((await elements(html,'.updates img'))[0]).toMatchObject({src:'/api/v1/media/cover',loading:'lazy',alt:''});
 expect((await elements(html,'.updates a[href^="/read/"]'))[0].href).toBe('/read/release');
 expect(await elements(html,'a a, a button')).toHaveLength(0);
});

test('home retains a working release feed when catalogue fails and keeps the selected retry',async()=>{
 const html=render(HomeDiscovery,{props:{available:false,shelf:'new',language:'ru'}}).body;
 expect(html).toContain('Каталог не загрузился');
 expect(await elements(html,'a.retry-link[href="/?shelf=new"]')).toHaveLength(1);
 expect(await elements(html,'.updates a[href="/read/release"]')).toHaveLength(1);
 expect(await elements(html,'.cover-score')).toHaveLength(0);
});

test('home shows image links first and real navigation rather than inert filter tabs',async()=>{
 const html=render(HomeDiscovery,{props:{shelf:'completed',language:'ru'}}).body;
 expect(html.indexOf('class="cover-shelf')).toBeLessThan(html.indexOf('class="updates'));
 expect((await elements(html,'.shelf-tabs a[aria-current]'))[0]).toMatchObject({href:'/?shelf=completed','aria-current':'page'});
 expect(await elements(html,'.cover-shelf a[href="/titles/example"] img')).toHaveLength(1);
 expect(await elements(html,'a a, a button')).toHaveLength(0);
 expect(await elements(render(HomeDiscovery,{props:{empty:true}}).body,'.cover-shelf')).toHaveLength(0);
});

test('cover scores are shown only when real votes exist',async()=>{
 expect(await elements(render(Discovery).body,'.cover-score')).toHaveLength(0);
 const rated=render(Discovery,{props:{ratingCount:3}}).body;
 expect(await elements(rated,'.cover-score')).toHaveLength(1);
 expect(rated).toContain('8.25');
});

test('global search submits to catalogue and preserves locale-specific accessible labels',async()=>{
 for(const language of ['kk','ru'] as const){
  const html=render(Discovery,{props:{language}}).body;
  expect((await elements(html,'dialog form'))[0]).toMatchObject({action:'/catalog',method:'GET',role:'search'});
  expect((await elements(html,'dialog input[name="q"]'))[0]).toMatchObject({type:'search',maxlength:'80',required:''});
  expect(html).toContain(language==='kk'?'Тайтл іздеу':'Поиск тайтла');
 }
});
