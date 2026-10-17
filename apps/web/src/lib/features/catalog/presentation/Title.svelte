<script lang="ts">
 import type {Snippet} from 'svelte';
 import type {TitleDetail,Chapter,ResultPage,ChapterFilter} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {catalogNames} from './messages';
 import ChapterRow from './ChapterRow.svelte';
 import TitleMetadata from './TitleMetadata.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import type {IconName} from '$lib/shared/presentation/ui/types';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
 let {title,chapters,nextURL,library,discussion,report,firstChapter=null,chapterFilter={},tab='chapters'}:{firstChapter?:Chapter|null;chapterFilter?:ChapterFilter;title:TitleDetail;chapters:ResultPage<Chapter>;nextURL:string|null;library?:Snippet;discussion?:Snippet;report?:Snippet;tab?:string}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const base=$derived('/titles/'+encodeURIComponent(title.slug));
 const active=$derived(['about','chapters','comments','stats'].includes(tab)?tab:'chapters');
 const tabs=$derived<{id:string;label:string;icon:IconName}[]>([
  {id:'about',label:text($locale,'Тайтл туралы','О тайтле'),icon:'help'},
  {id:'chapters',label:text($locale,'Тараулар','Главы'),icon:'menu'},
  {id:'comments',label:text($locale,'Пікірлер','Комментарии'),icon:'message'},
  {id:'stats',label:text($locale,'Статистика','Статистика'),icon:'chart'}
 ]);
 let shareNotice=$state('');
 async function share(){
  const url=new URL(base,window.location.origin).href;
  try{if(navigator.share)await navigator.share({title:title.name,url});else{await navigator.clipboard.writeText(url);shareNotice=text($locale,'Сілтеме көшірілді','Ссылка скопирована');}}
  catch(error){if(!(error instanceof Error&&error.name==='AbortError'))shareNotice=text($locale,'Көшіру мүмкін болмады. Браузердегі сілтемені көшір.','Не удалось скопировать. Скопируй адрес из браузера.');}
 }
</script>
<svelte:head><title>{title.name} — Qmanga</title><meta name="description" content={title.description.slice(0,160)}/></svelte:head>
<a class="back" href="/catalog"><Icon name="left" size={16}/>{text($locale,'Каталог','Каталог')}</a>
<div class="title-layout">
 <aside class="title-aside">
  <div class="cover-wrap">{#if title.coverURL}<img class="title-cover" src={title.coverURL} alt={title.name} width="280" height="400" fetchpriority="high"/>{:else}<div class="cover-empty"><Icon name="file" size={48}/><span>{text($locale,'Мұқаба жоқ','Нет обложки')}</span></div>{/if}</div>
  <div class="title-actions">
   {#if firstChapter}<a class="button primary read-button" href={firstChapter.releases.length===1?'/read/'+encodeURIComponent(firstChapter.releases[0].id):base+'?tab=chapters#chapters'}><Icon name="play" size={16}/>{text($locale,'Оқу','Читать')}</a>{:else}<button class="button" disabled>{text($locale,'Әзірге тарау жоқ','Глав пока нет')}</button>{/if}
   {#if library}<details class="library-choice"><summary><Icon name="bookmark" size={18}/>{text($locale,'Тізімге қосу','Добавить в список')}</summary><div class="library-panel">{@render library()}</div></details>{/if}
   <button class="quiet-action" type="button" onclick={share}><Icon name="share" size={18}/>{text($locale,'Бөлісу','Поделиться')}</button>
   {#if report}<div class="report-slot">{@render report()}</div>{/if}
   {#if shareNotice}<p role="status" class="muted">{shareNotice}</p>{/if}
  </div>
  <div class="sidebar-metadata"><TitleMetadata {title}/></div>
 </aside>
 <header class="title-heading">
  <div class="name-line"><div><h1>{title.name}</h1><p class="subtitle">{#if title.publication?.year}<span>{title.publication.year}</span>{/if}<a href={'/catalog?kind='+title.kind}>{names.kind[title.kind]}</a>{#if title.demo}<span class="demo">{text($locale,'Үлгі','Пример')}</span>{/if}</p></div><a class="rating" href={base+'?tab=comments#title-content'} aria-label={text($locale,'Бағалау','Оценить')}><strong>{title.ratingCount>0?title.score.toFixed(2):'—'}</strong><span>{title.ratingCount} {text($locale,'баға','оценок')}</span><span class="rate-label">{text($locale,'Бағалау','Оценить')}</span></a></div>
  {#if title.genres.length}<div class="title-genres">{#each title.genres as genre}<a href={'/catalog?genre='+encodeURIComponent(genre.slug)}>{genre.name}</a>{/each}</div>{/if}
  {#if title.description}<p class="synopsis">{title.description}</p>{/if}
 </header>
 <div class="title-main">
  <dl class="title-facts">
   <div><Icon name="file"/><div><dt>{text($locale,'Тайтл күйі','Статус тайтла')}</dt><dd><a href={'/catalog?original_status='+title.originalStatus}>{names.original[title.originalStatus]}</a></dd></div></div>
   <div><Icon name="users"/><div><dt>{text($locale,'Аударма күйі','Статус перевода')}</dt><dd><a href={'/catalog?translation_status='+title.translationStatus}>{names.translation[title.translationStatus]}</a></dd></div></div>
   <div><Icon name="menu"/><div><dt>{text($locale,'Тараулар','Главы')}</dt><dd>{title.chapterCount??'—'}</dd></div></div>
   <div><Icon name="catalog"/><div><dt>{text($locale,'Формат','Формат')}</dt><dd>{title.publication?.formats.length?title.publication.formats.join(', '):text($locale,'Белгісіз','Не указан')}</dd></div></div>
   <div><Icon name="eye"/><div><dt>{text($locale,'Қаралымдар','Просмотры')}</dt><dd><span title={text($locale,'Қаралым саны әзірге есептелмейді','Учёт просмотров пока не ведётся')}>—</span></dd></div></div>
  </dl>
  <nav class="title-tabs" aria-label={text($locale,'Тайтл бөлімдері','Разделы тайтла')}>{#each tabs as item}<a href={base+'?tab='+item.id+'#title-content'} aria-current={active===item.id?'page':undefined}><Icon name={item.icon} size={18}/>{item.label}</a>{/each}</nav>
  <div id="title-content" class="title-content">
   {#if active==='chapters'}
    <section id="chapters" aria-label={text($locale,'Тараулар','Главы')}>
     <form class="chapter-search" action={base+'#chapters'} method="GET"><input type="hidden" name="tab" value="chapters"/><label class="search-field"><Icon name="search" size={18}/><span class="sr-only">{text($locale,'Тарау іздеу','Найти главу')}</span><input name="q" type="search" maxlength="80" value={chapterFilter.query||''} placeholder={text($locale,'Тарау нөмірі немесе атауы','Номер или название главы')}/></label><select name="sort" value={chapterFilter.sort||'asc'} aria-label={text($locale,'Тарау реті','Порядок глав')}><option value="asc">{text($locale,'Басынан','Сначала старые')}</option><option value="desc">{text($locale,'Соңғысынан','Сначала новые')}</option></select><button class="button subtle" type="submit">{text($locale,'Іздеу','Найти')}</button></form>
     <div class="chapter-list">{#each chapters.items as chapter (chapter.id)}<ChapterRow title={(chapter.volume?text($locale,chapter.volume+'-том · ','Том '+chapter.volume+' · '):'')+text($locale,chapter.label+'-тарау','Глава '+chapter.label)+(chapter.name?' · '+chapter.name:'')} releases={chapter.releases} readLabel={text($locale,'Оқу','Читать')}/>{:else}<StateMessage kind="empty" title={chapterFilter.query?text($locale,'Тарау табылмады','Главы не найдены'):text($locale,'Жарияланған тарау жоқ','Опубликованных глав пока нет')}/>{/each}</div>
     {#if nextURL}<Pager label={text($locale,'Тарау беттері','Страницы глав')} next={{label:text($locale,'Келесі тараулар','Следующие главы'),href:nextURL}}/>{/if}
    </section>
   {:else if active==='about'}
    <section class="about"><h2>{text($locale,'Сипаттама','Описание')}</h2><p>{title.description||text($locale,'Сипаттама әлі қосылмаған','Описание пока не добавлено')}</p><div class="mobile-metadata"><TitleMetadata {title}/></div></section>
   {:else if active==='comments'}
    {#if discussion}{@render discussion()}{/if}
   {:else}
    <section class="statistics"><h2>{text($locale,'Статистика','Статистика')}</h2><dl><div><dt>{text($locale,'Жарияланған тараулар','Опубликованные главы')}</dt><dd>{title.chapterCount??'—'}</dd></div><div><dt>{text($locale,'Бағалар','Оценки')}</dt><dd>{title.ratingCount}</dd></div><div><dt>{text($locale,'Оқырмандар','Читатели')}</dt><dd>{title.readerCount??'—'}</dd></div></dl><p class="muted">{text($locale,'Оқырмандар — оқу орны сақталған аккаунттар. Қонақтағы оқу мен қаралымдар есепке кірмейді.','Читатели — аккаунты с сохранённым прогрессом. Гостевое чтение и просмотры сюда не входят.')}</p></section>
   {/if}
  </div>
 </div>
</div>
<style>
 .back{display:inline-flex;align-items:center;gap:4px;min-height:44px;color:var(--muted);font-size:13px;text-decoration:none;margin-bottom:8px}.title-layout{display:grid;grid-template-columns:100px minmax(0,1fr);gap:20px 14px;align-items:start}.title-aside{display:contents}.cover-wrap{grid-column:1;grid-row:1;min-width:0}.title-cover,.cover-empty{width:100%;aspect-ratio:7/10;object-fit:cover;border-radius:10px;height:auto;display:block}.cover-empty{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;background:var(--panel);color:var(--muted);font-size:12px;text-align:center}.title-heading{grid-column:2;grid-row:1;min-width:0}.name-line{display:flex;align-items:start;justify-content:space-between;gap:12px}h1{font-size:20px;line-height:1.25;margin:0;overflow-wrap:anywhere}.subtitle{display:flex;gap:8px;flex-wrap:wrap;margin:8px 0 0;font-size:13px;color:var(--muted)}.subtitle a{color:inherit;text-decoration:none}.demo{font-size:11px;border:1px solid var(--line);border-radius:4px;padding:0 4px}.rating{display:none;flex:none;align-items:center;text-decoration:none;flex-direction:column;font-size:12px;color:var(--muted);min-height:44px}.rating strong{color:var(--accent);font-size:21px;line-height:1.3}.rate-label{margin-top:4px;padding:6px 10px;border-radius:6px;background:var(--panel);color:var(--text)}.title-genres{display:flex;flex-wrap:wrap;gap:6px;margin-top:14px}.title-genres a{display:inline-flex;align-items:center;min-height:44px;padding:4px 9px;border-radius:7px;background:var(--panel);font-size:12px;font-weight:550;text-decoration:none;color:var(--text)}.synopsis{display:none}.title-actions{grid-column:1/-1;display:grid;grid-template-columns:1fr 1fr;gap:8px;min-width:0}.title-actions>.button{width:100%;display:flex;gap:8px;align-items:center;justify-content:center}.library-choice{min-width:0}.library-choice summary{display:flex;align-items:center;justify-content:center;gap:8px;min-height:44px;padding:8px;border:1px solid var(--line);border-radius:8px;background:var(--panel);font-size:13px;font-weight:600;cursor:pointer;list-style:none}.library-choice summary::-webkit-details-marker{display:none}.library-panel{padding:10px;border:1px solid var(--line);border-radius:8px;margin-top:8px}.library-choice[open]{grid-column:1/-1}.quiet-action{display:flex;justify-content:center;align-items:center;gap:8px;min-height:44px;border:0;background:none;color:var(--muted);font-size:13px}.report-slot{min-width:0}.report-slot:has(:global(details[open])){grid-column:1/-1}.report-slot :global(details){margin:0}.report-slot :global(summary){justify-content:center}.report-slot :global(details[open]){min-width:0}.title-main{grid-column:1/-1;min-width:0}.title-facts{margin:0 0 20px;padding:14px;background:var(--panel);border:1px solid var(--line);border-radius:12px;display:grid;grid-template-columns:1fr 1fr;gap:16px 12px}.title-facts>div{display:flex;gap:9px;align-items:center;min-width:0}.title-facts>div>:global(svg){color:var(--muted);width:18px}.title-facts dt{color:var(--muted);font-size:12px;margin-bottom:4px}.title-facts dd{margin:0;font-size:13px;font-weight:600;overflow-wrap:anywhere}.title-facts a{color:var(--text);text-decoration:none;display:inline-flex;align-items:center;min-height:44px;margin-block:-10px}.title-tabs{display:flex;gap:4px;overflow-x:auto;scrollbar-width:thin;padding-bottom:8px;margin-bottom:12px}.title-tabs a{flex:none;display:flex;align-items:center;gap:7px;min-height:44px;padding:8px 10px;border:1px solid transparent;border-radius:9px;text-decoration:none;color:var(--muted);font-size:13px;font-weight:600}.title-tabs a[aria-current]{border-color:var(--accent);color:var(--text);background:var(--panel)}#title-content,#chapters{scroll-margin-top:80px}.chapter-search{display:flex;flex-wrap:wrap;gap:8px;margin-bottom:16px}.search-field{display:flex;align-items:center;gap:8px;border:1px solid var(--line);border-radius:9px;background:var(--panel);padding-left:12px;flex:1 1 100%;min-width:0;color:var(--muted)}.search-field input{border:0;background:transparent;min-width:0;width:100%;padding-left:0}.chapter-search select{flex:1;width:auto;min-width:0}.chapter-list{display:grid;gap:8px}.about p{white-space:pre-wrap;line-height:1.7;overflow-wrap:anywhere}.about h2,.statistics h2{font-size:19px;margin:0 0 16px}.statistics dl{display:grid;gap:16px;grid-template-columns:repeat(3,minmax(0,1fr));margin:0 0 20px}.statistics dt{font-size:12px;color:var(--muted)}.statistics dd{font-size:24px;margin:8px 0}.statistics p{font-size:13px}.sidebar-metadata{display:none}.mobile-metadata{margin-top:24px}
 @media(min-width:600px){.title-layout{grid-template-columns:150px minmax(0,1fr);gap:24px}.rating{display:flex}.title-facts{grid-template-columns:repeat(3,minmax(0,1fr))}h1{font-size:26px}}
 @media(min-width:960px){.title-layout{grid-template-columns:250px minmax(0,1fr);grid-template-rows:auto 1fr;gap:24px}.title-aside{display:flex;flex-direction:column;gap:10px;grid-column:1;grid-row:1/3;min-width:0}.title-heading{grid-column:2;grid-row:1}.title-main{grid-column:2;grid-row:2}.title-actions{display:flex;flex-direction:column;width:100%;gap:6px}.cover-wrap{width:100%}.sidebar-metadata{display:block;margin-top:10px;width:100%}.synopsis{display:-webkit-box;line-clamp:2;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin:18px 0 0;line-height:1.6;font-size:14px}.title-genres{margin-top:18px}.title-facts{grid-template-columns:1.2fr 1.2fr .65fr 1.2fr .7fr;gap:12px;padding:16px}.search-field{flex:1 1 220px}.chapter-search select{flex:0 1 170px}.mobile-metadata{display:none}}
</style>