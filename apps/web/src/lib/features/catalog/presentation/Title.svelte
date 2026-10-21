<script lang="ts">
 import TitleBanner from './TitleBanner.svelte';
 import TitleTags from './TitleTags.svelte';
 import AgeRating from './AgeRating.svelte';
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
 let {title,chapters,nextURL,library,discussion,report,rating,firstChapter=null,chapterFilter={},tab='chapters',libraryLabel=''}:{libraryLabel?:string;firstChapter?:Chapter|null;chapterFilter?:ChapterFilter;title:TitleDetail;chapters:ResultPage<Chapter>;nextURL:string|null;library?:Snippet;discussion?:Snippet;report?:Snippet;rating?:Snippet;tab?:string}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const taxonomy=$derived([...(title.categories||[]).map(item=>({key:'category-'+item.id,href:'/catalog?category='+encodeURIComponent(item.slug),label:$locale==='ru'&&item.nameRu?item.nameRu:item.name})),...(title.collections||[]).map(item=>({key:'collection-'+item.id,href:'/catalog?collection='+encodeURIComponent(item.slug),label:item.name})),...title.genres.map(item=>({key:'genre-'+item.id,href:'/catalog?genre='+encodeURIComponent(item.slug),label:$locale==='ru'&&item.nameRu?item.nameRu:item.name}))]);
 const base=$derived('/titles/'+encodeURIComponent(title.slug));
 const active=$derived(['about','chapters','comments'].includes(tab)?tab:'chapters');
 const tabs=$derived<{id:string;label:string;icon:IconName;count?:number}[]>([
  {id:'about',label:text($locale,'Тайтл туралы','О тайтле'),icon:'help'},
  {id:'chapters',label:text($locale,'Тараулар','Главы'),icon:'menu',count:title.chapterCount},
  {id:'comments',label:text($locale,'Пікірлер','Комментарии'),icon:'message',count:title.commentCount}
 ]);
 let statsDialog=$state<HTMLDialogElement|undefined>();
 let libraryDialog=$state<HTMLDialogElement|undefined>(),reportDialog=$state<HTMLDialogElement|undefined>();
 const maxRating=$derived(Math.max(1,...(title.ratingDistribution||[])));
 let slideDirection=$state(1);
 let descriptionExpanded=$state(false),descriptionCanExpand=$state(false);
 function observeDescription(node:HTMLParagraphElement,body:string){
  let currentBody=body,disposed=false;
  const measure=()=>{if(disposed)return;const lineHeight=parseFloat(getComputedStyle(node).lineHeight);descriptionCanExpand=!!currentBody&&node.scrollHeight>Math.ceil(lineHeight*3)+1;if(!descriptionCanExpand)descriptionExpanded=false;};
  const observer=new ResizeObserver(measure);observer.observe(node);
  document.fonts.addEventListener('loadingdone',measure);void document.fonts.ready.then(measure);
  measure();
  return {update(value:string){currentBody=value;queueMicrotask(measure);},destroy(){disposed=true;observer.disconnect();document.fonts.removeEventListener('loadingdone',measure);}};
 }
 $effect(()=>{title.id;descriptionExpanded=false;});
 let shareNotice=$state('');
 async function share(){
  const url=new URL(base,window.location.origin).href;
  try{if(navigator.share)await navigator.share({title:title.name,url});else{await navigator.clipboard.writeText(url);shareNotice=text($locale,'Сілтеме көшірілді','Ссылка скопирована');}}
  catch(error){if(!(error instanceof Error&&error.name==='AbortError'))shareNotice=text($locale,'Көшіру мүмкін болмады. Браузердегі сілтемені көшір.','Не удалось скопировать. Скопируй адрес из браузера.');}
 }
</script>
<svelte:head><title>{title.name} — Qmanga</title><meta name="description" content={title.description.slice(0,160)}/></svelte:head>
<a class="back" href="/catalog"><Icon name="left" size={16}/>{text($locale,'Каталог','Каталог')}</a>
<TitleBanner bannerURL={title.bannerURL||''} coverURL={title.coverURL}/>
<div class="title-layout" class:about-view={active==='about'} class:with-banner={!!(title.bannerURL||title.coverURL)}>
 <header class="title-heading">
  <div class="name-line"><div><h1>{title.name}</h1><p class="subtitle"><AgeRating value={title.ageRating}/>{#if title.publication?.year}<span>{title.publication.year}</span>{/if}<a href={'/catalog?kind='+title.kind}>{names.kind[title.kind]}</a>{#if title.demo}<span class="demo">{text($locale,'Үлгі','Пример')}</span>{/if}</p></div>{#if rating}{@render rating()}{/if}</div>
  <TitleTags items={taxonomy} id={'title-tags-'+title.id}/>
  {#if title.description}<p class="synopsis">{title.description}</p>{/if}
 </header>
 <aside class="title-aside">
  <div class="cover-wrap">{#if title.coverURL}<img class="title-cover" src={title.coverURL} alt={title.name} width="280" height="400" fetchpriority="high"/>{:else}<div class="cover-empty"><Icon name="file" size={48}/><span>{text($locale,'Мұқаба жоқ','Нет обложки')}</span></div>{/if}</div>
  <div class="title-actions">
   {#if firstChapter}<a class="button primary read-button" href={firstChapter.releases.length===1?'/read/'+encodeURIComponent(firstChapter.releases[0].id):base+'?tab=chapters#chapters'}><Icon name="play" size={16}/>{text($locale,'Оқу','Читать')}</a>{:else}<button class="button" disabled>{text($locale,'Әзірге тарау жоқ','Глав пока нет')}</button>{/if}
   {#if library}<button class="button subtle library-trigger" type="button" onclick={()=>libraryDialog?.showModal()}><Icon name="bookmark" size={18}/>{libraryLabel||text($locale,'Тізімге қосу','Добавить в список')}</button><dialog bind:this={libraryDialog} class="action-dialog" aria-labelledby="library-title"><div class="dialog-heading"><h2 id="library-title">{text($locale,'Менің тізімім','Мой список')}</h2><button class="button subtle" type="button" onclick={()=>libraryDialog?.close()} aria-label={text($locale,'Жабу','Закрыть')}>×</button></div>{@render library()}</dialog>{/if}
   <button class="quiet-action" type="button" onclick={share}><Icon name="share" size={18}/>{text($locale,'Бөлісу','Поделиться')}</button>
   {#if report}<button class="quiet-action" type="button" onclick={()=>reportDialog?.showModal()}><Icon name="flag" size={18}/>{text($locale,'Мәселе туралы хабарлау','Сообщить о проблеме')}</button><dialog bind:this={reportDialog} class="action-dialog" aria-labelledby="report-title"><div class="dialog-heading"><h2 id="report-title">{text($locale,'Мәселе туралы хабарлау','Сообщить о проблеме')}</h2><button class="button subtle" type="button" onclick={()=>reportDialog?.close()} aria-label={text($locale,'Жабу','Закрыть')}>×</button></div>{@render report()}</dialog>{/if}
   {#if shareNotice}<p role="status" class="muted">{shareNotice}</p>{/if}
  </div>
  {#if active==='about'}<div class="title-information"><TitleMetadata {title} onStatistics={()=>statsDialog?.showModal()}/></div>{/if}
 </aside>

  <nav class="title-tabs" style:--active-tab={tabs.findIndex(item=>item.id===active)} aria-label={text($locale,'Тайтл бөлімдері','Разделы тайтла')}>{#each tabs as item}<a data-sveltekit-noscroll href={base+'?tab='+item.id} onclick={()=>slideDirection=tabs.findIndex(value=>value.id===item.id)>=tabs.findIndex(value=>value.id===active)?1:-1} aria-current={active===item.id?'page':undefined}><Icon name={item.icon} size={18}/><span>{item.label}</span>{#if item.count!==undefined}<span class="tab-count">{item.count}</span>{/if}</a>{/each}</nav>
 <div class="title-main">
  <dialog bind:this={statsDialog} class="stats-dialog" aria-labelledby="stats-title"><div class="stats-dialog-heading"><h2 id="stats-title">{text($locale,'Тайтл статистикасы','Статистика тайтла')}</h2><button class="button subtle" type="button" onclick={()=>statsDialog?.close()} aria-label={text($locale,'Жабу','Закрыть')}>×</button></div><dl class="stats-summary"><div><dt>{text($locale,'Бағалар','Оценки')}</dt><dd>{title.ratingCount}</dd></div><div><dt>{text($locale,'Тараулар','Главы')}</dt><dd>{title.chapterCount??'—'}</dd></div><div><dt>{text($locale,'Оқырмандар','Читатели')}</dt><dd>{title.readerCount??'—'}</dd></div></dl><h3>{text($locale,'Бағалардың таралуы','Распределение оценок')}</h3>{#if title.ratingCount>0}<div class="rating-bars">{#each [10,9,8,7,6,5,4,3,2,1] as value}<div class="rating-bar"><span>{value}</span><span class="rating-bar-track"><span style:width={(100*(title.ratingDistribution?.[value-1]||0)/maxRating)+'%'}></span></span><span>{title.ratingDistribution?.[value-1]||0}</span></div>{/each}</div>{:else}<p class="muted">{text($locale,'Әзірге баға жоқ','Оценок пока нет')}</p>{/if}<p class="muted stats-explainer">{text($locale,'Оқырмандар — оқу орны сақталған аккаунттар. Қонақтағы оқу мен қаралымдар есепке кірмейді.','Читатели — аккаунты с сохранённым прогрессом. Гостевое чтение и просмотры сюда не входят.')}</p></dialog>

  <div id="title-content" class="title-content">
   {#key active}<div class="tab-panel" style:--tab-shift={(slideDirection*12)+'px'}>
   {#if active==='chapters'}
    <section id="chapters" aria-label={text($locale,'Тараулар','Главы')}>
     <form class="chapter-search" action={base+'#chapters'} method="GET"><input type="hidden" name="tab" value="chapters"/><label class="search-field"><input type="search" name="q" value={chapterFilter.query||''} placeholder={text($locale,'Тарау нөмірі немесе атауы','Номер или название главы')} aria-label={text($locale,'Тарау іздеу','Поиск главы')} maxlength="80"/></label><button class="button subtle" type="submit" name="sort" value={chapterFilter.sort||'asc'} aria-label={text($locale,'Іздеу','Найти')} title={text($locale,'Іздеу','Найти')}><Icon name="search"/></button><button class="button subtle" type="submit" name="sort" value={chapterFilter.sort==='desc'?'asc':'desc'} aria-label={chapterFilter.sort==='desc'?text($locale,'Басынан','Сначала старые'):text($locale,'Соңғысынан','Сначала новые')} title={chapterFilter.sort==='desc'?text($locale,'Басынан','Сначала старые'):text($locale,'Соңғысынан','Сначала новые')}><Icon name={chapterFilter.sort==='desc'?'down':'up'}/></button></form>
     <div class="chapter-list">{#each chapters.items as chapter (chapter.id)}<ChapterRow title={(chapter.volume?text($locale,chapter.volume+'-том · ','Том '+chapter.volume+' · '):'')+text($locale,chapter.label+'-тарау','Глава '+chapter.label)+(chapter.name?' · '+chapter.name:'')} releases={chapter.releases} readLabel={text($locale,'Оқу','Читать')}/>{:else}<StateMessage kind="empty" title={chapterFilter.query?text($locale,'Тарау табылмады','Главы не найдены'):text($locale,'Жарияланған тарау жоқ','Опубликованных глав пока нет')}/>{/each}</div>
     {#if nextURL}<Pager label={text($locale,'Тарау беттері','Страницы глав')} next={{label:text($locale,'Келесі тараулар','Следующие главы'),href:nextURL}}/>{/if}
    </section>
   {:else if active==='about'}
    <section class="about"><h2>{text($locale,'Сипаттама','Описание')}</h2><p id="title-description" use:observeDescription={title.description} class:description-collapsed={!descriptionExpanded}>{title.description||text($locale,'Сипаттама әлі қосылмаған','Описание пока не добавлено')}</p>{#if descriptionCanExpand}<button class="description-toggle" type="button" aria-expanded={descriptionExpanded} aria-controls="title-description" onclick={()=>descriptionExpanded=!descriptionExpanded}>{descriptionExpanded?text($locale,'Жию','Свернуть'):text($locale,'Толығырақ','Подробнее')}<Icon name={descriptionExpanded?'up':'down'} size={16}/></button>{/if}</section>
   {:else if active==='comments'}
    {#if discussion}{@render discussion()}{/if}
   {/if}
   </div>{/key}
  </div>
 </div>
</div>
<style>
 .back{display:inline-flex;align-items:center;gap:4px;min-height:44px;color:var(--muted);font-size:13px;text-decoration:none;margin-bottom:8px}.title-layout{display:grid;grid-template-columns:100px minmax(0,1fr);gap:20px 14px;align-items:start}.title-aside{display:contents}.cover-wrap{grid-column:1;grid-row:1;min-width:0}.title-cover,.cover-empty{width:100%;aspect-ratio:7/10;object-fit:cover;border-radius:10px;height:auto;display:block}.cover-empty{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;background:var(--panel);color:var(--muted);font-size:12px;text-align:center}.title-heading{grid-column:2;grid-row:1;min-width:0}.name-line{display:flex;align-items:start;justify-content:space-between;gap:12px}h1{font-size:20px;line-height:1.25;margin:0;overflow-wrap:anywhere}.subtitle{display:flex;gap:8px;flex-wrap:wrap;margin:8px 0 0;font-size:13px;color:var(--muted)}.subtitle a{color:inherit;text-decoration:none}.demo{font-size:11px;border:1px solid var(--line);border-radius:4px;padding:0 4px}.synopsis{display:none}.title-actions{grid-column:1/-1;display:grid;grid-template-columns:1fr 1fr;gap:8px;min-width:0}.title-actions>.button{width:100%;display:flex;gap:8px;align-items:center;justify-content:center}.quiet-action{display:flex;justify-content:center;align-items:center;gap:8px;min-height:44px;border:0;background:none;color:var(--muted);font-size:13px}.title-main{grid-column:1/-1;min-width:0}.title-tabs{display:flex;gap:4px;overflow-x:auto;scrollbar-width:thin;padding-bottom:8px;margin-bottom:12px}.title-tabs a{flex:none;display:flex;align-items:center;gap:7px;min-height:44px;padding:8px 10px;border:1px solid transparent;border-radius:9px;text-decoration:none;color:var(--muted);font-size:13px;font-weight:600}.title-tabs a[aria-current]{border-color:var(--accent);color:var(--text);background:var(--panel)}#title-content,#chapters{scroll-margin-top:80px}.chapter-search{display:flex;flex-wrap:nowrap;gap:8px;margin-bottom:16px}.search-field{display:flex;align-items:center;gap:8px;border:1px solid var(--line);border-radius:9px;background:var(--panel);padding-inline:12px;flex:1;min-width:0;color:var(--muted)}.search-field input{border:0;background:transparent;min-width:0;width:100%;padding-left:0}.chapter-search button{min-width:44px;flex:none;padding:10px}.tab-count{font-size:12px;color:var(--muted);font-variant-numeric:tabular-nums}.chapter-list{display:grid;gap:8px}.about p{white-space:pre-wrap;line-height:1.7;overflow-wrap:anywhere}.about h2{font-size:19px;margin:0 0 16px}
 @media(min-width:600px){.title-layout{grid-template-columns:150px minmax(0,1fr);gap:24px}h1{font-size:26px}}
 @media(min-width:960px){.title-layout{grid-template-columns:250px minmax(0,1fr);grid-template-rows:auto 1fr;gap:24px}.title-aside{display:flex;flex-direction:column;gap:10px;grid-column:1;grid-row:1/3;min-width:0}.title-heading{grid-column:2;grid-row:1}.title-main{grid-column:2;grid-row:2}.title-actions{display:flex;flex-direction:column;width:100%;gap:6px}.cover-wrap{width:100%}.synopsis{display:-webkit-box;line-clamp:2;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden;margin:18px 0 0;line-height:1.6;font-size:14px}.search-field{flex:1 1 220px}}
 .name-line{flex-wrap:wrap}.name-line>div{min-width:0}.name-line :global(.rating-trigger){flex-direction:column;align-items:flex-start;gap:0;max-width:100%;min-width:0}.name-line :global(.rate-label){margin-top:4px}@media(min-width:600px){.name-line{flex-wrap:nowrap}.name-line :global(.rating-trigger){align-items:center}}.stats-dialog{width:min(420px,calc(100vw - 24px));max-height:min(650px,calc(100dvh - 24px));overflow:auto;padding:20px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);box-shadow:0 24px 72px #0008}.stats-dialog::backdrop{background:#0009}.stats-dialog-heading{display:flex;align-items:center;justify-content:space-between;gap:12px}.stats-dialog h2{font-size:19px;margin:0}.stats-dialog h3{font-size:15px;margin:20px 0 10px}.stats-dialog-heading button{min-width:44px;font-size:23px}.stats-summary{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin:16px 0}.stats-summary div{border:1px solid var(--line);border-radius:8px;padding:10px}.stats-summary dt{font-size:11px;color:var(--muted)}.stats-summary dd{font-size:20px;font-weight:700;margin:4px 0 0}.rating-bars{display:grid;gap:7px}.rating-bar{display:grid;grid-template-columns:20px minmax(0,1fr) 36px;gap:8px;align-items:center;font-size:12px}.rating-bar>span:last-child{text-align:right;color:var(--muted)}.rating-bar-track{display:block;height:9px;border-radius:9px;background:var(--raised);overflow:hidden}.rating-bar-track>span{display:block;height:100%;border-radius:inherit;background:var(--accent)}.stats-explainer{font-size:12px;line-height:1.5;margin-bottom:0}
.library-trigger{width:100%;display:flex;align-items:center;justify-content:center;gap:8px}.action-dialog{width:min(520px,calc(100vw - 24px));max-height:calc(100dvh - 24px);overflow:auto;padding:20px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);box-shadow:0 24px 72px #0008}.action-dialog::backdrop{background:#0009}.dialog-heading{display:flex;justify-content:space-between;align-items:center;gap:12px}.dialog-heading h2{font-size:19px;margin:0}.dialog-heading button{min-width:44px;font-size:23px}
@media(min-width:960px){.title-layout.with-banner{position:relative;margin-top:-44px}.with-banner .title-heading{padding-top:56px}.with-banner .cover-wrap{border:4px solid var(--bg);border-radius:12px;overflow:hidden}.with-banner .title-aside{padding-left:12px}}

 @media(max-width:767px){
  .title-layout{grid-template-columns:100px minmax(0,1fr);gap:14px 12px}
  .cover-wrap{grid-column:1;grid-row:1;width:100%;align-self:start}
  .title-heading{grid-column:2;grid-row:1;text-align:left}
  .name-line{flex-direction:column;align-items:flex-start;gap:10px}
  h1{font-size:21px;line-height:1.3}.subtitle{margin-top:6px;gap:5px}
  .name-line :global(.rating-trigger){flex-direction:row;align-items:center;justify-content:flex-start;flex-wrap:wrap;gap:6px}
  .name-line :global(.rate-label){margin-top:0;padding:4px 6px}


  .title-actions{grid-column:1/-1;grid-row:2;gap:8px}
  .title-actions>.read-button,.title-actions>.library-trigger,.title-actions>button:disabled{grid-column:1/-1}
  .quiet-action{padding:6px;line-height:1.4;overflow-wrap:anywhere}
  .title-tabs{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));overflow:visible;gap:2px}
  .title-tabs a{min-width:0;justify-content:center;flex-wrap:wrap;gap:4px;padding:8px 4px;font-size:12px;overflow-wrap:anywhere}
  .title-tabs :global(svg){display:none}
 }

 .about p.description-collapsed{display:-webkit-box;-webkit-box-orient:vertical;-webkit-line-clamp:3;line-clamp:3;overflow:hidden}
 .description-toggle{display:inline-flex;align-items:center;gap:6px;min-height:44px;padding:4px 0;border:0;background:none;color:var(--accent);font:inherit;font-weight:600;cursor:pointer}
 .description-toggle:hover{text-decoration:underline}.description-toggle:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
 .title-tabs{position:relative;isolation:isolate;display:grid;grid-template-columns:repeat(3,minmax(0,1fr));overflow:visible;gap:0}
 .title-tabs::before{content:'';position:absolute;inset:0 auto 0 0;width:calc(100% / 3);border:1px solid var(--accent);border-radius:9px;background:var(--panel);transform:translateX(calc(var(--active-tab)*100%));transition:transform 180ms ease;pointer-events:none;z-index:-1}
 .title-tabs a{justify-content:center;min-width:0}.title-tabs a[aria-current]{border-color:transparent;background:none}
 .tab-panel{animation:tab-enter 160ms ease-out}
 @keyframes tab-enter{from{opacity:.6;transform:translateX(var(--tab-shift))}to{opacity:1;transform:translateX(0)}}
 @media(prefers-reduced-motion:reduce){.title-tabs::before{transition:none}.tab-panel{animation:none}}
 @media(max-width:959px){.title-layout:not(.about-view) .title-main{grid-row:4}}
 /* One information panel: sidebar on desktop, below navigation on small screens. */
 .title-information{grid-column:1/-1;grid-row:4;min-width:0}
 .title-tabs{grid-column:1/-1;grid-row:3;margin:0;padding:0}
 .title-main{grid-row:5}
 @media(max-width:767px){
  .title-tabs{grid-row:3}
  .title-information{grid-row:4}
  .title-main{grid-row:5}
 }
 @media(min-width:960px){
  .title-layout{grid-template-columns:200px minmax(0,1fr);grid-template-rows:auto auto 1fr;gap:16px 24px}
  .title-aside{grid-row:1/4;gap:12px}
  .title-information{width:100%}
  .title-tabs{grid-column:2;grid-row:2}
  .title-main{grid-column:2;grid-row:3}
  .synopsis{display:none}
 }
</style>
