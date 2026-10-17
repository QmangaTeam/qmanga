import { expect, test } from 'bun:test';
import Content from './Content.fixture.svelte';
import { render, elements } from './markup';

test('content semantics preserve headings, state priority and field labels', async () => {
 const html = render(Content).body;
 expect(await elements(html, 'h1')).toHaveLength(1);
 expect(await elements(html, 'h2')).toHaveLength(1);
 expect(await elements(html, 'h3')).toHaveLength(1);
 expect(await elements(html, '[data-case="empty"] [role]')).toHaveLength(0);
 expect(await elements(html, '[data-case="error"] [role="alert"]')).toHaveLength(1);
 expect(await elements(html, '[role="status"]')).toHaveLength(1);
 expect((await elements(html, 'label'))[0].for).toBe('quality');
});

test('pager uses real links and disables busy callbacks', async () => {
 const html = render(Content).body;
 expect((await elements(html, 'nav a'))[0].href).toBe('/catalog?cursor=test');
 expect((await elements(html, 'nav button'))[0]).toMatchObject({type:'button',disabled:'','aria-busy':'true'});
});
