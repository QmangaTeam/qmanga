import {expect,test} from 'bun:test';
import {render,elements} from './markup';
import Menu from './Menu.fixture.svelte';

test('menu destinations are icon-backed native links, never nested buttons',async()=>{
 const html=render(Menu).body;
 for(const destination of ['/saved','/settings','/help','/reports']){
  expect(await elements(html,'a[href="'+destination+'"] svg')).toHaveLength(1);
 }
 expect(await elements(html,'a button,a a')).toHaveLength(0);
 expect(html).not.toContain('private@example.test');
 expect(await elements(html,'a[href="/users/account"]')).toHaveLength(1);
 expect(await elements(html,'a[href="/catalog"],a[href="/library"],a[href="/notifications"],a[href^="/studio"]')).toHaveLength(0);
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
 expect(await elements(admin,'a[href="/admin"]')).toHaveLength(1);
 expect(await elements(admin,'a[href^="/admin/"],a[href="/moderation"]')).toHaveLength(0);
});

test('unverified staff have no management destinations',async()=>{
 for(const role of ['admin','moderator'] as const){
  const html=render(Menu,{props:{role,verified:false}}).body;
  expect(await elements(html,'a[href="/admin"],a[href="/moderation"],a[href^="/studio"]')).toHaveLength(0);
 }
});
