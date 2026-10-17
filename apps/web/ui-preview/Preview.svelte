<script lang="ts">
 import { onMount, onDestroy } from 'svelte';
 import Button from '../src/lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '../src/lib/shared/presentation/ui/ActionLink.svelte';
 import IconButton from '../src/lib/shared/presentation/ui/IconButton.svelte';
 import Field from '../src/lib/shared/presentation/ui/Field.svelte';
 import Badge from '../src/lib/shared/presentation/ui/Badge.svelte';
 import SectionHeading from '../src/lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '../src/lib/shared/presentation/ui/StateMessage.svelte';
 import SettingsRow from '../src/lib/shared/presentation/ui/SettingsRow.svelte';
 import Pager from '../src/lib/shared/presentation/ui/Pager.svelte';
 import CoverCard from '../src/lib/features/catalog/presentation/CoverCard.svelte';
 import ChapterRow from '../src/lib/features/catalog/presentation/ChapterRow.svelte';
 import LibraryRow from '../src/lib/features/reading/presentation/LibraryRow.svelte';

 let language = $state<'kk' | 'ru'>('kk');
 let theme = $state<'dark' | 'light'>('dark');
 let count = $state(0), busy = $state(false), title = $state(''), notice = $state(false);
 let timeout: ReturnType<typeof setTimeout> | undefined;
 const copy = $derived(language === 'kk' ? {
  title: 'Компоненттер', note: 'Дизайнның жұмыс үлгілері. Аккаунт пен манга деректерін өзгертпейді.',
  actions: 'Батырмалар', primary: 'Сақтау', secondary: 'Бас тарту', ghost: 'Қосымша', danger: 'Тізімнен алу',
  fields: 'Өрістер', name: 'Тайтл атауы', hint: 'Қазақша атауын жаз.', comment: 'Пікір', placeholder: 'Осы жерге жаз',
  invalid: 'Сілтеме', error: 'Тек латын әріптері мен дефис қолдан.', disabled: 'Өңдеу қолжетімсіз',
  save: 'Сынап көру', saved: 'Дайын. Бұл тек интерфейс үлгісі.', waiting: 'Сақталуда…',
  light: 'Жарық', dark: 'Қараңғы', counter: 'Басу саны', reset: 'Санақты тазарту', link: 'Өрістерге өту',
  form: 'Қарапайым форма', send: 'Жіберу', long: 'Атауы ұзын тайтлды кітапханаға сақтап, кейін оқуды жалғастыру'
 } : {
  title: 'Компоненты', note: 'Рабочие примеры дизайна. Не изменяют аккаунт или данные манги.',
  actions: 'Кнопки', primary: 'Сохранить', secondary: 'Отмена', ghost: 'Ещё', danger: 'Убрать из списка',
  fields: 'Поля', name: 'Название тайтла', hint: 'Укажи название на казахском.', comment: 'Комментарий', placeholder: 'Напиши здесь',
  invalid: 'Короткая ссылка', error: 'Используй только латинские буквы и дефисы.', disabled: 'Редактирование недоступно',
  save: 'Проверить', saved: 'Готово. Это только пример интерфейса.', waiting: 'Сохранение…',
  light: 'Светлая', dark: 'Тёмная', counter: 'Нажатий', reset: 'Сбросить счётчик', link: 'К полям',
  form: 'Обычная форма', send: 'Отправить', long: 'Сохранить тайтл с длинным названием в библиотеку и продолжить чтение позже'
 });

 function changeTheme(value: 'dark' | 'light') { theme = value; document.documentElement.dataset.theme = value; }
 function changeLanguage(value: 'kk' | 'ru') { language = value; document.documentElement.lang = value; }
 function save() { busy = true; notice = false; count++; timeout = setTimeout(() => { busy = false; notice = true; }, 1200); }
 onMount(() => { document.documentElement.dataset.theme = theme; });
 onDestroy(() => { if (timeout) clearTimeout(timeout); });
</script>

<header><a href="#top" class="wordmark">qmanga<span>.</span></a><span class="preview-label">UI preview</span></header>
<main id="top">
 <div class="preview-heading"><h1>{copy.title}</h1><p>{copy.note}</p></div>
 <div class="preview-switches">
  <div role="group" aria-label="Қазақша / Русский"><Button variant="ghost" aria-pressed={language === 'kk'} onclick={() => changeLanguage('kk')}>Қазақша</Button><Button variant="ghost" aria-pressed={language === 'ru'} onclick={() => changeLanguage('ru')}>Русский</Button></div>
  <div role="group" aria-label={language === 'kk' ? 'Тақырып' : 'Тема'}><Button variant="ghost" aria-pressed={theme === 'dark'} onclick={() => changeTheme('dark')}>{copy.dark}</Button><Button variant="ghost" aria-pressed={theme === 'light'} onclick={() => changeTheme('light')}>{copy.light}</Button></div>
 </div>
 <section aria-labelledby="actions-heading">
  <h2 id="actions-heading">{copy.actions}</h2>
  <div class="actions"><Button variant="primary" onclick={() => count++}>{copy.primary}</Button><Button onclick={() => count++}>{copy.secondary}</Button><Button variant="ghost" onclick={() => count++}>{copy.ghost}</Button><Button variant="danger" onclick={() => count++}>{copy.danger}</Button><IconButton name="close" label={copy.reset} onclick={() => count = 0}/></div>
  <p class="counter" role="status">{copy.counter}: {count}</p>
  <div class="actions"><Button variant="primary" {busy} onclick={save}>{busy ? copy.waiting : copy.save}</Button><Button disabled>{copy.disabled}</Button><ActionLink href="#fields">{copy.link}</ActionLink></div>
  <p class="feedback" role="status">{notice ? copy.saved : ''}</p>
  <div class="long-control"><Button onclick={() => count++}>{copy.long}</Button></div>
 </section>
 <section id="fields" aria-labelledby="fields-heading">
  <h2 id="fields-heading">{copy.fields}</h2>
  <div class="field-grid">
   <form onsubmit={event => { event.preventDefault(); notice = true; }} aria-label={copy.form}>
    <Field id="preview-title" label={copy.name} hint={copy.hint} required>{#snippet children(control)}<input {...control} bind:value={title} autocomplete="off" placeholder={copy.placeholder}/>{/snippet}</Field>
    <Field id="preview-comment" label={copy.comment}>{#snippet children(control)}<textarea {...control} placeholder={copy.placeholder}></textarea>{/snippet}</Field>
    <Button type="submit" variant="primary">{copy.send}</Button>
   </form>
   <div class="field-samples">
    <Field id="preview-invalid" label={copy.invalid} error={copy.error}>{#snippet children(control)}<input {...control} value="үлгі" autocomplete="off"/>{/snippet}</Field>
    <Field id="preview-disabled" label={copy.disabled} disabled>{#snippet children(control)}<select {...control}><option>Қазақша</option></select>{/snippet}</Field>
   </div>
  </div>
 </section>
 <section>
  <SectionHeading title={language==='kk'?'Контент үлгілері':'Примеры контента'} description={language==='kk'?'Төмендегі атаулар — тек үлгі.':'Названия ниже — демонстрационные.'}/>
  <Badge label={language==='kk'?'Оқып жүрмін':'Читаю'} tone="accent"/>
  <div class="sample-covers"><CoverCard href="#top" name={language==='kk'?'Ұзақ атауы бар демонстрациялық тайтл':'Демонстрационный тайтл с длинным названием'} metadata="Манхва" detail="12" fallback={language==='kk'?'Мұқаба жоқ':'Нет обложки'}/><CoverCard href="#top" name={language==='kk'?'Тағы бір үлгі':'Ещё один пример'} metadata="Манга" fallback={language==='kk'?'Мұқаба жоқ':'Нет обложки'}/></div>
  <ChapterRow title={language==='kk'?'1-тарау · Үлгі':'Глава 1 · Пример'} releases={[]} readLabel={language==='kk'?'Оқу':'Читать'}/>
  <LibraryRow name={language==='kk'?'Кітапханадағы үлгі':'Пример в библиотеке'} href="#top" detail={language==='kk'?'Оқу әлі басталмаған':'Чтение ещё не начато'} continueLabel={language==='kk'?'Жалғастыру':'Продолжить'}/>
  <SettingsRow controlId="preview-quality" label={language==='kk'?'Сапа':'Качество'}><select id="preview-quality"><option>{language==='kk'?'Авто':'Авто'}</option></select></SettingsRow>
  <StateMessage kind="empty" title={language==='kk'?'Әзірге бос':'Пока пусто'} description={language==='kk'?'Қосылған тайтлдар осы жерде болады.':'Здесь появятся добавленные тайтлы.'}/>
  <StateMessage kind="error" title={language==='kk'?'Жүктеу мүмкін болмады':'Не удалось загрузить'}>{#snippet actions()}<Button onclick={()=>count++}>{language==='kk'?'Қайталау':'Повторить'}</Button>{/snippet}</StateMessage>
  <Pager label={language==='kk'?'Беттер':'Страницы'} previous={{label:language==='kk'?'Артқа':'Назад',action:()=>count++,disabled:true}} next={{label:language==='kk'?'Келесі':'Далее',href:'#top'}}/>
 </section>
</main>

<style>
 :global(body) { margin: 0; background: var(--bg); color: var(--text); font-size: 16px; line-height: 1.5; }
 :global(*) { box-sizing: border-box; }
 :global(:focus-visible) { outline: 2px solid var(--accent); outline-offset: 3px; }
 :global([aria-pressed="true"]) { background: var(--raised); color: var(--accent); }
 :global(html) { scroll-padding-top: 72px; }
 header { height: 56px; display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 0 16px; border-bottom: 1px solid var(--line); }
 .wordmark { min-height: 44px; display: flex; align-items: center; color: var(--text); font-size: 23px; font-weight: 750; letter-spacing: -1px; text-decoration: none; }
 .wordmark span { color: var(--accent); }
 .preview-label { color: var(--muted); font-size: 13px; }
 main { max-width: 1100px; margin: auto; padding: 24px 16px 64px; }
 h1, h2, p { margin: 0; }
 h1 { font-size: 24px; line-height: 1.3; font-weight: 650; }
 h2 { font-size: 18px; line-height: 1.4; font-weight: 650; margin-bottom: 16px; }
 .preview-heading p { max-width: 620px; color: var(--muted); margin-top: 8px; font-size: 14px; }
 .preview-switches, .preview-switches>div, .actions { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
 .preview-switches { justify-content: space-between; margin-top: 24px; padding-bottom: 16px; }
 section { padding: 24px 0; border-top: 1px solid var(--line); }
 .counter { color: var(--muted); font-size: 14px; margin: 12px 0 16px; }
 .feedback { min-height: 24px; font-size: 14px; margin: 12px 0; }
 .long-control { max-width: 360px; }
 .sample-covers { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:16px; max-width:400px; margin:24px 0; }
 .field-grid, form, .field-samples { display: grid; align-content: start; gap: 24px; min-width: 0; }
 form :global(.ui-action) { justify-self: start; }
 @media (min-width: 768px) { header { padding-inline: 24px; } main { padding-inline: 24px; } h1 { font-size: 28px; } .field-grid { grid-template-columns: 1fr 1fr; gap: 32px; } }
 @media (prefers-reduced-motion: reduce) { :global(*), :global(*::before), :global(*::after) { animation: none !important; transition: none !important; scroll-behavior: auto !important; } }
</style>
