import {expect,test} from 'bun:test';
import {render,elements} from './markup';
import Branding from './Branding.fixture.svelte';
test('auth wordmark has accessible single-Q branding and preserves form destination',async()=>{
 const html=render(Branding).body;
 expect(await elements(html,'.auth-brand')).toMatchObject([{href:'/', 'aria-label':'Qmanga'}]);
 expect(await elements(html,'.auth-brand img')).toMatchObject([{src:'/favicon.png',alt:''}]);
 expect(await elements(html,'form')).toMatchObject([{action:'/login?returnTo=%2Flibrary',method:'POST'}]);
});
test('mobile destinations retain names and notification badge with Home centered',async()=>{
 for(const currentPath of ['/','/notifications']){
  const html=render(Branding,{props:{currentPath}}).body;
  const links=await elements(html,'.mobile-navigation a');
  expect(links.map(link=>link.href)).toEqual(['/catalog','/','/library','/notifications']);
  expect(links[1]['aria-label']).toBe('Басты бет');
  expect(await elements(html,'.mobile-navigation>button')).toMatchObject([{'aria-haspopup':'dialog'}]);
  expect(await elements(html,'.mobile-navigation a[aria-current="page"]')).toMatchObject([{href:currentPath}]);
  expect(await elements(html,'.mobile-navigation a[href="/"] img')).toMatchObject([{src:'/favicon.png',alt:''}]);
  expect(await elements(html,'.mobile-navigation a[href="/"] svg')).toHaveLength(0);
  expect(await elements(html,'.mobile-navigation a[href="/notifications"]')).toMatchObject([{'aria-label':'Хабарлар: 12'}]);
  expect(await elements(html,'.mobile-navigation b')).toHaveLength(1);
 }
});
