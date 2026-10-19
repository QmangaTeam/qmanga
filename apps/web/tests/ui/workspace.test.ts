import {test,expect} from 'bun:test';
import Navigation from '../../src/lib/shared/presentation/ui/WorkspaceNavigation.svelte';
import {render,elements} from './markup';
const items=[{href:'/studio/projects',label:'Жобалар'},{href:'/studio/requests',label:'Өтінімдер'}];
test('sidebar navigation exposes a named collapsed mobile toggle and preserves nested active routes',async()=>{
 const html=render(Navigation,{props:{items,currentPath:'/studio/projects/example',label:'Кабинет',sidebar:true,menuLabel:'Бөлімдер'}}).body;
 expect(await elements(html,'button[aria-expanded="false"][aria-controls="workspace-sections"]')).toHaveLength(1);
 expect(html).toContain('Бөлімдер');
 expect((await elements(html,'nav a[aria-current="page"]'))[0].href).toBe('/studio/projects');
 expect(await elements(html,'nav#workspace-sections a')).toHaveLength(2);
});
test('ordinary workspace tabs retain their layout without a disclosure button',async()=>{
 const html=render(Navigation,{props:{items,currentPath:'/studio/requests',label:'Кабинет'}}).body;
 expect(await elements(html,'button')).toHaveLength(0);
 expect(await elements(html,'nav a')).toHaveLength(2);
});
