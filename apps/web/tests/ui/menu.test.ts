import {expect,test} from 'bun:test';
import {render,elements} from './markup';
import Menu from './Menu.fixture.svelte';

test('menu destinations are icon-backed native links, never nested buttons',async()=>{
 const html=render(Menu).body;
 for(const destination of ['/library','/saved','/notifications','/studio/projects','/studio/requests','/studio/teams','/reports','/help']){
  expect(await elements(html,'a[href="'+destination+'"] svg')).toHaveLength(1);
 }
 expect(await elements(html,'a button,a a')).toHaveLength(0);
 expect(html).not.toContain('private@example.test');
});

test('language choices expose the selected locale without submitting a form',async()=>{
 for(const language of ['kk','ru'] as const){
  const html=render(Menu,{props:{language}}).body;
  const selected=await elements(html,'.language button[aria-pressed="true"]');
  expect(selected).toHaveLength(1);
  expect(selected[0]).toMatchObject({type:'button','data-value':language});
  expect(await elements(html,'.language select')).toHaveLength(0);
 }
});

test('menu retains guest return path and staff-only destinations',async()=>{
 const guest=render(Menu,{props:{guest:true}}).body;
 expect(await elements(guest,'a[href="/login?returnTo=%2Fmenu"]')).toHaveLength(1);
 expect(await elements(guest,'a[href^="/admin"],a[href="/moderation"]')).toHaveLength(0);
 const reader=render(Menu).body;
 expect(await elements(reader,'a[href^="/admin"],a[href="/moderation"]')).toHaveLength(0);
 const moderator=render(Menu,{props:{role:'moderator'}}).body;
 expect(await elements(moderator,'a[href="/moderation"]')).toHaveLength(1);
 expect(await elements(moderator,'a[href^="/admin"]')).toHaveLength(0);
 const admin=render(Menu,{props:{role:'admin'}}).body;
 for(const destination of ['/moderation','/admin','/admin/requests','/admin/catalog']){
  expect(await elements(admin,'a[href="'+destination+'"]')).toHaveLength(1);
 }
});
