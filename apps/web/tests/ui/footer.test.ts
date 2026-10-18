import {test,expect} from 'bun:test';
import {render,elements} from './markup';
import Footer from './Footer.fixture.svelte';
test('compact footer owns useful native links, authentic translator help and build identity',async()=>{
 for(const language of ['kk','ru'] as const){const html=render(Footer,{props:{language}}).body;
 expect(await elements(html,'section')).toHaveLength(4);
 for(const href of ['/settings','/reports','/help','/rights','/studio/teams','/studio/requests','/studio/projects','/help#publishing','/legal/terms','/legal/copyright','/legal/privacy'])expect(await elements(html,`a[href="${href}"]`)).toHaveLength(1);
 expect(html).toContain('abcdef1');expect(html).toContain('© 2026 Qmanga');expect(html).not.toContain('Senkuro');expect(html).not.toContain('оферта');
 }
});
