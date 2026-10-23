<script lang="ts">
 import {tick,type Snippet} from 'svelte';
 import type {ResultPage,Summary,ReleaseUpdate} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import CoverShelf from './CoverShelf.svelte';
 import Updates from './Updates.svelte';
 let {mostRead,updates,top,topKind='manhwa',continuation,statuses={},hasContinuation=false,topRequested=false}:{hasContinuation?:boolean;topRequested?:boolean;statuses?:Record<string,string>;mostRead:{available:boolean;page:ResultPage<Summary>};updates:{available:boolean;page:ResultPage<ReleaseUpdate>};top:{available:boolean;page:ResultPage<Summary>};topKind?:string;continuation?:Snippet}=$props();
 const {locale}=useLocale();
 let updateLimit=$state(6);
 let updateRevealFrom=$state(Infinity);
 let updateList=$state<HTMLDivElement>();
 async function showMoreUpdates(){
  const firstNew=updateLimit;
  updateRevealFrom=firstNew;
  updateLimit+=6;
  await tick();
  updateList?.querySelectorAll<HTMLAnchorElement>('.chapter')[firstNew]?.focus({preventScroll:true});
 }
 $effect(()=>{updates.page;updateLimit=6;updateRevealFrom=Infinity;});
 const tabs=$derived([{value:'manhwa',label:text($locale,'Топ манхва','Топ манхв')},{value:'manga',label:text($locale,'Топ манга','Топ манг')},{value:'manhua',label:text($locale,'Топ маньхуа','Топ маньхуа')}]);
 const ranking=$derived(topRequested?top:mostRead);
 const rankingLabel=$derived(topRequested?tabs.find(tab=>tab.value===topKind)?.label||tabs[0].label:text($locale,'Ең көп оқылатын','Самое читаемое'));
 const sparse=$derived(ranking.available&&ranking.page.items.length<4);
</script>
{#snippet continueSection()}
 <section aria-labelledby="continue-heading">
  <div class="section-title"><h2 id="continue-heading">{text($locale,'Оқуды жалғастыру','Продолжить чтение')}</h2><a class="section-link" href="/library" aria-label={text($locale,'Кітапхана','Библиотека')}><Icon name="bookmark" size={18}/></a></div>
  {#if continuation}{@render continuation()}{/if}
 </section>
{/snippet}
<div class="discovery-home" class:sparse>
 <section id="tops" aria-labelledby="browse-heading">
  <div class="section-title"><h1 id="browse-heading">{text($locale,'Үздік тайтлдар','Топ тайтлов')}</h1><a class="section-link" href="/catalog" aria-label="Каталог"><Icon name="catalog" size={18}/></a></div>
  <nav class="shelf-tabs" aria-label={text($locale,'Рейтингті таңдау','Выбор рейтинга')}>
   <a href="/#tops" data-sveltekit-noscroll data-sveltekit-keepfocus data-sveltekit-preload-data="hover" aria-current={!topRequested?'page':undefined}>{text($locale,'Ең көп оқылатын','Самое читаемое')}</a>
   {#each tabs as tab}<a href={'/?top='+tab.value+'#tops'} data-sveltekit-noscroll data-sveltekit-keepfocus data-sveltekit-preload-data="hover" aria-current={topRequested&&topKind===tab.value?'page':undefined}>{tab.label}</a>{/each}
  </nav>
  {#key topRequested?topKind:'most-read'}
  <div class="ranking-panel">
  {#if !ranking.available}<StateMessage kind="error" title={text($locale,'Рейтинг жүктелмеді','Рейтинг не загрузился')}>{#snippet actions()}<a class="retry-link" href={topRequested?'/?top='+topKind+'#tops':'/'}>{text($locale,'Қайталау','Повторить')}</a>{/snippet}</StateMessage>
  {:else if !ranking.page.items.length}<StateMessage kind="empty" title={text($locale,'Бұл бөлімде әзірге тайтл жоқ','В этом разделе пока нет тайтлов')}/>
  {:else}<CoverShelf {statuses} items={ranking.page.items} label={rankingLabel} eager/>{/if}
  </div>
  {/key}
 </section>
 {@render continueSection()}
 <section aria-labelledby="updates-heading">
  <div class="section-title"><h2 id="updates-heading">{text($locale,'Соңғы жаңартулар','Последние обновления')}</h2><a class="section-link" href="/updates" aria-label={text($locale,'Барлық жаңартулар','Все обновления')}><Icon name="right" size={18}/></a></div>
  {#if updates.available}<div id="home-update-list" bind:this={updateList}><Updates items={updates.page.items.slice(0,updateLimit)} revealFrom={updateRevealFrom} compact/></div>
   <div class="updates-footer">
    <span class="update-count" role="status">{text($locale,'Көрсетілді: ','Показано: ')}{Math.min(updateLimit,updates.page.items.length)} / {updates.page.items.length}{updates.page.nextCursor?'+':''}</span>
    {#if updateLimit<updates.page.items.length}<button class="button subtle" aria-controls="home-update-list" onclick={showMoreUpdates}>{text($locale,'Тағы жүктеу','Загрузить ещё')}<Icon name="down" size={16}/></button>
    {:else if updates.page.nextCursor}<a class="button subtle" href="/updates">{text($locale,'Барлық жаңартулар','Все обновления')}<Icon name="right" size={16}/></a>{/if}
   </div>{:else}<StateMessage kind="error" title={text($locale,'Жаңартулар жүктелмеді','Не удалось загрузить обновления')}/>{/if}
 </section>

</div>
<style>
 .ranking-panel{animation:ranking-enter 180ms ease-out;min-width:0}
 @keyframes ranking-enter{from{opacity:.35;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
 .shelf-tabs a{transition:background-color 160ms ease,color 160ms ease}
 @media(prefers-reduced-motion:reduce){.ranking-panel{animation:none}.shelf-tabs a{transition:none}}
 :global(html[data-motion=reduce]) .ranking-panel{animation:none}
 :global(html[data-motion=reduce]) .shelf-tabs a{transition:none}
 .updates-footer{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;margin-top:12px}.update-count{font-size:12px;color:var(--muted)}
 .discovery-home{display:grid;gap:28px;min-width:0}.discovery-home.sparse{gap:20px}.discovery-home>section{min-width:0}.section-title{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:8px}h1,h2{font-size:19px;font-weight:650;line-height:1.35;overflow-wrap:anywhere}.section-link{border:1px solid var(--line);border-radius:8px;background:var(--panel);padding:0 10px;display:flex;align-items:center;justify-content:center;min-height:44px;min-width:44px;color:var(--muted);flex-shrink:0}.section-link:hover{color:var(--text)}.shelf-tabs{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));max-width:100%;gap:2px;padding:3px;margin-bottom:12px;border:1px solid var(--line);border-radius:12px}.shelf-tabs a{min-height:44px;display:flex;align-items:center;justify-content:center;padding:0 4px;border-radius:8px;font-size:11px;font-weight:550;white-space:nowrap;color:var(--muted)}.shelf-tabs a[aria-current]{background:var(--raised);color:var(--text)}.shelf-tabs a:hover{color:var(--text)}.retry-link{min-height:44px;display:inline-flex;align-items:center;color:var(--accent)}#tops{scroll-margin-top:80px}@media(min-width:768px){.discovery-home{gap:36px}h1,h2{font-size:24px}.shelf-tabs{width:max-content;grid-template-columns:repeat(4,minmax(0,1fr))}.shelf-tabs a{font-size:14px;padding-inline:12px}}
</style>
