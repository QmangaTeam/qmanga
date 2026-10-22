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

test('workspace shell retains real destinations, single-Q brand and a hydration-safe mobile drawer',async()=>{
 const {default:Shell}=await import('./WorkspaceShell.fixture.svelte');
 const html=render(Shell,{props:{compact:true}}).body;
 expect(await elements(html,'main#content h1')).toHaveLength(1);
 expect((await elements(html,'.workspace-brand img'))[0].src).toBe('/favicon.png');
 expect(await elements(html,'.workspace-sidebar a[aria-current="page"]')).toHaveLength(1);
 expect(await elements(html,'button.workspace-menu[disabled][aria-haspopup="dialog"]')).toHaveLength(1);
 expect((await elements(html,'.workspace-catalog'))[0].href).toBe('/catalog');
 expect(html).toContain('3');expect(await elements(html,'.workspace-drawer nav a')).toHaveLength(2);
 expect(await elements(html,'.sidebar-bottom a[aria-label][title]')).toHaveLength(2);
});
