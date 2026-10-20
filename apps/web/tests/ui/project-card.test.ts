import {test,expect} from 'bun:test';
import Card from '../../src/lib/features/publishing/presentation/ProjectCard.svelte';
import {render,elements} from './markup';
test('project card retains its cover and distinguishes project chapters from title ratings',async()=>{
 const html=render(Card,{props:{item:{id:'project',titleName:'Тайтл',teamName:'Команда',coverURL:'/api/v1/media/cover',chapterCount:4,ratingCount:3,score:8.5}}}).body;
 expect((await elements(html,'img'))[0].src).toBe('/api/v1/media/cover');expect(await elements(html,'article dl dd')).toHaveLength(2);expect(html).toContain('4');expect(html).toContain('8.50');
 expect((await elements(html,'a.manage'))[0].href).toBe('/studio/projects/project');
});
test('missing card data and unrated titles do not invent zero votes or covers',async()=>{
 const html=render(Card,{props:{item:{id:'project',titleName:'Тайтл',teamName:'Команда',ratingCount:0,score:0}}}).body;
 expect(await elements(html,'img')).toHaveLength(0);expect(html).not.toContain('0.00');expect(html).toContain('—');
});
test('project action names the real revision and deep-links to it without guessing from counts',async()=>{
 const html=render(Card,{props:{item:{id:'project',titleName:'Тайтл',teamName:'Команда',nextRevisionId:'draft-id',nextRevisionState:'processing'}}}).body;
 expect((await elements(html,'a.manage'))[0].href).toBe('/studio/projects/project?revision=draft-id');
 expect(html).toContain('Өңделуде');
});
test('unknown work state keeps a neutral action and list return is explicit',async()=>{
 const html=render(Card,{props:{item:{id:'project',titleName:'Тайтл',teamName:'Команда',nextRevisionId:'revision',nextRevisionState:'future-state'},returnTo:'/studio/projects?q=Applied&pages=2'}}).body;
 expect(html).not.toContain('Тексеруге жіберу');expect((await elements(html,'a.manage'))[0].href).toBe('/studio/projects/project?revision=revision&amp;projectsReturn=%2Fstudio%2Fprojects%3Fq%3DApplied%26pages%3D2');
});
