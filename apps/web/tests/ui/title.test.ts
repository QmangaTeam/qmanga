import {test,expect} from 'bun:test';
import Title from './Title.fixture.svelte';
import {render,elements} from './markup';

test('title places actions and publication metadata beside a unified facts panel',async()=>{
 const html=render(Title).body;
 expect(await elements(html,'.title-aside a[href="/read/release"]')).toHaveLength(1);
 expect(await elements(html,'.title-information .metadata')).toHaveLength(1);
 expect(await elements(html,'.title-facts,.sidebar-metadata,.mobile-metadata')).toHaveLength(0);
 expect(await elements(html,'.metadata .summary-facts')).toHaveLength(1);
 expect(await elements(html,'.metadata .statistics')).toHaveLength(1);
 expect(html).toContain('185');
 expect(html).toContain('Publisher');
 expect(html).toContain('Корея');
 expect((await elements(html,'a[href="https://example.com/title"]'))[0].rel).toContain('noopener');
 expect(await elements(html,'h1')).toHaveLength(1);
});
test('title tabs retain real chapters and expose only the selected content',async()=>{
 const chapters=render(Title).body;
 expect((await elements(chapters,'.title-tabs a[aria-current]'))[0].href).toContain('tab=chapters');
 expect(await elements(chapters,'input[name="tab"][value="chapters"]')).toHaveLength(1);
 expect(chapters).not.toContain('Тестовые комментарии');
 const comments=render(Title,{props:{tab:'comments'}}).body;
 expect(comments).toContain('Тестовые комментарии');
 expect(await elements(comments,'.chapter-search')).toHaveLength(0);
 expect(await elements(chapters,'a a, a button')).toHaveLength(0);
});

test('title distinguishes explicit age ratings without classifying unknown titles',async()=>{
 const adult=render(Title,{props:{ageRating:18}}).body;
 expect(await elements(adult,'.age-rating.adult')).toHaveLength(1);
 expect(adult).toContain('18+');
 expect(await elements(render(Title).body,'.age-rating')).toHaveLength(0);
});

 test('title tabs show complete chapter and comment counts and chapter search uses icon actions',async()=>{
  const html=render(Title).body;
  expect(await elements(html,'.title-tabs .tab-count')).toHaveLength(2);
  const counts:string[]=[];
  await new HTMLRewriter().on('.title-tabs .tab-count',{text(chunk){counts.push(chunk.text);}}).transform(new Response(html)).text();
  expect(counts.join('')).toBe('18512');
  expect(await elements(html,'.chapter-search select')).toHaveLength(0);
  expect(await elements(html,'.chapter-search button[name="sort"][value="asc"]')).toHaveLength(1);
  expect(await elements(html,'.chapter-search button[name="sort"][value="desc"]')).toHaveLength(1);
  expect(await elements(html,'.search-field svg')).toHaveLength(0);
 });

test('title uses a wide cover fallback and an explicit banner when provided',async()=>{
 const fallback=render(Title).body;
 expect((await elements(fallback,'.title-banner img'))[0].src).toBe('/api/v1/media/cover');
 const explicit=render(Title,{props:{bannerURL:'/api/v1/media/banner'}}).body;
 expect((await elements(explicit,'.title-banner img'))[0].src).toBe('/api/v1/media/banner');
});

test('description preview has an accessible localized disclosure',async()=>{
 for(const language of ['kk','ru'] as const){
  const html=render(Title,{props:{tab:'about',language}}).body;
  const buttons=await elements(html,'.description-toggle');expect(buttons).toHaveLength(1);
  expect(buttons[0]['aria-expanded']).toBe('false');expect(buttons[0]['aria-controls']).toBe('title-description');
  expect(await elements(html,'#title-description.description-collapsed')).toHaveLength(1);
  expect(html).toContain(language==='kk'?'Толығырақ':'Подробнее');
 }
});

test('title tab links preserve scroll without a content fragment',async()=>{
 const html=render(Title).body;
 for(const link of await elements(html,'.title-tabs a')){expect(link.href).not.toContain('#');expect(link['data-sveltekit-noscroll']).toBeDefined();}
 expect(await elements(html,'.metadata .metadata-rail')).toHaveLength(1);
 expect(html).toContain('Издатель');
 expect(await elements(html,'.metadata .contributors')).toHaveLength(0);
 expect(await elements(html,'.title-information>.contributors')).toHaveLength(1);
});
