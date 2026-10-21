import {test,expect} from 'bun:test';
import Discovery from './Discovery.fixture.svelte';
import HomeDiscovery from './HomeDiscovery.fixture.svelte';
import {render,elements} from './markup';

test('home follows most-read, continuation, updates and type-top order',()=>{
 const html=render(HomeDiscovery,{props:{language:'ru'}}).body;
 expect(html).toContain('Самое читаемое');
 const mostRead=html.indexOf('id="browse-heading"');
 const continuation=html.indexOf('id="continue-heading"');
 const updates=html.indexOf('id="updates-heading"');
 const tops=html.indexOf('id="tops-heading"');
 expect(mostRead).toBeGreaterThan(-1);
 expect(continuation).toBeGreaterThan(mostRead);
 expect(updates).toBeGreaterThan(continuation);
 expect(tops).toBeGreaterThan(updates);
});

test('below-fold type tops do not eagerly load images on weak connections',async()=>{
 const html=render(HomeDiscovery,{props:{topFilled:true,topRequested:true}}).body;
 expect((await elements(html,'section[aria-labelledby="browse-heading"] img'))[0]).toMatchObject({loading:'eager',fetchpriority:'high'});
 expect((await elements(html,'#tops img'))[0]).toMatchObject({loading:'lazy',fetchpriority:'auto'});
});

test('update thumbnail opens its title while chapter opens the real release',async()=>{
 const html=render(Discovery).body;
 expect(await elements(html,'.updates a[href="/titles/example"] img')).toHaveLength(1);
 expect((await elements(html,'.updates img'))[0]).toMatchObject({src:'/api/v1/media/cover',loading:'lazy',alt:''});
 expect((await elements(html,'.updates a[href^="/read/"]'))[0].href).toBe('/read/release');
 expect(await elements(html,'a a, a button')).toHaveLength(0);
});

test('home retains a working release feed when most-read ranking fails',async()=>{
 const html=render(HomeDiscovery,{props:{available:false,language:'ru'}}).body;
 expect(html).toContain('Рейтинг не загрузился');
 expect(await elements(html,'a.retry-link[href="/"]')).toHaveLength(1);
 expect(await elements(html,'.updates a[href="/read/release"]')).toHaveLength(1);
 expect(await elements(html,'.cover-score')).toHaveLength(0);
});

test('home shows image links first and real navigation rather than inert filter tabs',async()=>{
 const html=render(HomeDiscovery,{props:{topKind:'manhua',language:'ru'}}).body;
 expect(html.indexOf('class="cover-shelf')).toBeLessThan(html.indexOf('class="updates'));
 expect((await elements(html,'.shelf-tabs a[aria-current]'))[0]).toMatchObject({href:'/?top=manhua#tops','aria-current':'page'});
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

test('header wordmark uses one Q symbol followed by manga in every theme',async()=>{
 const html=render(Discovery).body;
 expect((await elements(html,'header a.wordmark img'))).toHaveLength(1);
 expect((await elements(html,'header a.wordmark img'))[0]).toMatchObject({alt:'',src:'/favicon.png',width:'40',height:'40'});
 expect(html).toContain('manga');
 expect((await elements(html,'header a.wordmark'))[0]['aria-label']).toBe('Qmanga');
});

test('cover classification keeps native links and never labels unclassified content',async()=>{
 expect(await elements(render(Discovery).body,'.cover-card .age-rating')).toHaveLength(0);
 const html=render(Discovery,{props:{ageRating:18}}).body;
 expect(await elements(html,'.cover-card .age-rating.adult')).toHaveLength(1);
 expect(await elements(html,'a a, a button')).toHaveLength(0);
 expect(html).toContain('18+');
});


test('returning readers reach continuation first without removing discovery',()=>{
 const html=render(HomeDiscovery,{props:{hasContinuation:true}}).body;
 expect(html.indexOf('id="continue-heading"')).toBeLessThan(html.indexOf('id="browse-heading"'));
 expect(html).toContain('id="updates-heading"');
});

test('default repeated type rankings collapse while explicit type requests retain results',async()=>{
 const defaultPage=render(HomeDiscovery,{props:{topFilled:true}}).body;
 expect(await elements(defaultPage,'#tops .cover-card')).toHaveLength(0);
 expect(await elements(defaultPage,'.shelf-tabs a')).toHaveLength(3);
 const explicit=render(HomeDiscovery,{props:{topFilled:true,topRequested:true}}).body;
 expect(await elements(explicit,'#tops .cover-card')).toHaveLength(1);
});
