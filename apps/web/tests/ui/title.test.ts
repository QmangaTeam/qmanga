import {test,expect} from 'bun:test';
import Title from './Title.fixture.svelte';
import {render,elements} from './markup';

test('title places actions and publication metadata beside a unified facts panel',async()=>{
 const html=render(Title,{props:{tab:'about'}}).body;
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

test('description SSR keeps the preview without offering an unmeasured disclosure',async()=>{
 for(const language of ['kk','ru'] as const){
  const html=render(Title,{props:{tab:'about',language}}).body;
  expect(await elements(html,'.description-toggle')).toHaveLength(0);
  expect(await elements(html,'#title-description.description-collapsed')).toHaveLength(1);
  expect(html).toContain('Описание произведения');
 }
});

test('title tab links preserve scroll without a content fragment',async()=>{
 const html=render(Title,{props:{tab:'about'}}).body;
 for(const link of await elements(html,'.title-tabs a')){expect(link.href).not.toContain('#');expect(link['data-sveltekit-noscroll']).toBeDefined();expect(link['data-sveltekit-keepfocus']).toBeDefined();expect(link['data-sveltekit-preload-data']).toBe('hover');}
 expect(await elements(html,'.metadata .metadata-rail')).toHaveLength(1);
 expect(html).toContain('Издатель');
 expect(await elements(html,'.metadata .contributors')).toHaveLength(0);
 expect(await elements(html,'.title-information .contributors')).toHaveLength(1);
});

test('title metadata uses field icons, status dots and a team avatar without repeating chapter totals',async()=>{
 const html=render(Title,{props:{tab:'about'}}).body;
 expect(await elements(html,'.summary-facts>div')).toHaveLength(8);
 expect(await elements(html,'.summary-facts dt svg')).toHaveLength(8);
 expect(await elements(html,'.status-dot')).toHaveLength(2);
 expect(await elements(html,'.contributors .team-avatar img')).toHaveLength(1);
 expect((await elements(html,'.team-avatar img'))[0].src).toBe('/media/avatars/team/team');
});


test('title without artwork keeps a placeholder and working reading destination',async()=>{
 const html=render(Title,{props:{coverURL:''}}).body;
 expect(await elements(html,'.cover-empty')).toHaveLength(1);
 expect(await elements(html,'.title-banner')).toHaveLength(0);
 expect((await elements(html,'.read-button'))[0].href).toBe('/read/release');
});


test('title information is exclusive to About, leaving chapter and comment views focused',async()=>{
 for(const tab of ['chapters','comments']){const html=render(Title,{props:{tab}}).body;expect(await elements(html,'.title-information,.metadata,.contributors')).toHaveLength(0);}
 expect(await elements(render(Title,{props:{tab:'about'}}).body,'.title-information')).toHaveLength(1);
});

 test('title team selector keeps exact team URLs and selected filter through search',async()=>{
  const html=render(Title,{props:{teamCount:10,selectedTeam:'team-7'}}).body;
  expect(await elements(html,'.translation-team')).toHaveLength(11);
  expect((await elements(html,'.translation-team[aria-current]'))[0].href).toContain('team=team-7');
  expect((await elements(html,'.translation-team[aria-current]'))[0].href).not.toContain('cursor=');
  expect(await elements(html,'input[name="team"][value="team-7"]')).toHaveLength(1);
  expect(await elements(html,'.translation-avatar img')).toHaveLength(10);
  expect(await elements(render(Title).body,'.translation-teams')).toHaveLength(0);
 });

 test('title separates genres, tags and collections into their own linked groups',async()=>{
  const html=render(Title,{props:{tab:'about',genreCount:3,categoryCount:4,collectionCount:2}}).body;
  expect(await elements(html,'section[aria-labelledby="title-genres-heading"] a[href^="/catalog?genre="]')).toHaveLength(3);
  expect(await elements(html,'section[aria-labelledby="title-tags-heading"] a.tag[href^="/catalog?category="]')).toHaveLength(4);
  expect(await elements(html,'section[aria-labelledby="title-collections-heading"] a[href^="/catalog?collection="]')).toHaveLength(2);
  expect(await elements(html,'.about-taxonomy')).toHaveLength(3);
 });
