import {test,expect} from 'bun:test';
import {render} from './markup';
import LocaleFixture from './Locale.fixture.svelte';
import {noticeText,copy,uiLabels} from '../../src/lib/shared/presentation/copy';
test('help explains offline limits and reporting in both UI languages',()=>{
 const russian=render(LocaleFixture,{props:{language:'ru',screen:'help'}}).body;
 const kazakh=render(LocaleFixture,{props:{language:'kk',screen:'help'}}).body;
 expect(russian).toContain('Помощь и правила');
 expect(russian).toContain('Сообщить о проблеме');
 expect(russian).toContain('браузер может удалить');
 expect(kazakh).toContain('Көмек және ережелер');
 expect(kazakh).toContain('Мәселе туралы хабарлау');
 expect(kazakh).toContain('href="/saved"');
});
test('SSR locale contexts do not leak between renders',()=>{
 const russian=render(LocaleFixture,{props:{language:'ru'}}).body;
 const kazakh=render(LocaleFixture,{props:{language:'kk'}}).body;
 expect(russian).toContain('Тайтлы не найдены');
 expect(russian).not.toContain('Тайтл табылмады');
 expect(kazakh).toContain('Тайтл табылмады');
 expect(kazakh).not.toContain('Тайтлы не найдены');
 expect(render(LocaleFixture,{props:{language:'ru'}}).body).toContain('Фильтры');
 expect(kazakh).toContain('Сүзгілер');
 expect(kazakh).not.toContain('Фильтры');
});
test('notices translate at render time and never expose unknown backend text',()=>{
 expect(noticeText('ru','progress_pending')).toContain('синхронизации');
 expect(noticeText('kk','progress_pending')).toContain('синхрондау');
 expect(noticeText('ru','private internal exception')).not.toContain('private');
 expect(noticeText('kk','none')).toBe('');
 for(const key of Object.keys(uiLabels) as (keyof typeof uiLabels)[]) {
  expect(copy('kk',key).trim().length).toBeGreaterThan(0);
  expect(copy('ru',key).trim().length).toBeGreaterThan(0);
 }
});
