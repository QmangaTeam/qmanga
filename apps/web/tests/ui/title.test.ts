import {test,expect} from 'bun:test';
import Title from './Title.fixture.svelte';
import {render,elements} from './markup';

test('title places actions and publication metadata beside a unified facts panel',async()=>{
 const html=render(Title).body;
 expect(await elements(html,'.title-aside a[href="/read/release"]')).toHaveLength(1);
 expect(await elements(html,'.title-facts')).toHaveLength(1);
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
