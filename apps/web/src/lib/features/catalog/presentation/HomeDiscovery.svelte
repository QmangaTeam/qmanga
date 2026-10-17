<script lang="ts">
 import type {Snippet} from 'svelte';
 import type {ResultPage,Summary,ReleaseUpdate} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import CoverShelf from './CoverShelf.svelte';
 import Updates from './Updates.svelte';
 let {mostRead,updates,top,topKind='manhwa',continuation}:{mostRead:{available:boolean;page:ResultPage<Summary>};updates:{available:boolean;page:ResultPage<ReleaseUpdate>};top:{available:boolean;page:ResultPage<Summary>};topKind?:string;continuation?:Snippet}=$props();
 const {locale}=useLocale();
 const tabs=$derived([{value:'manhwa',label:text($locale,'Топ манхва','Топ манхв')},{value:'manga',label:text($locale,'Топ манга','Топ манг')},{value:'manhua',label:text($locale,'Топ маньхуа','Топ маньхуа')}]);
</script>
<div class="discovery-home">
 <section aria-labelledby="browse-heading">
  <div class="section-title"><h1 id="browse-heading">{text($locale,'Ең көп оқылатын','Самое читаемое')}</h1><a class="section-link" href="/catalog" aria-label="Каталог"><Icon name="catalog" size={18}/></a></div>
  {#if !mostRead.available}<StateMessage kind="error" title={text($locale,'Рейтинг жүктелмеді','Рейтинг не загрузился')}>{#snippet actions()}<a class="retry-link" href="/">{text($locale,'Қайталау','Повторить')}</a>{/snippet}</StateMessage>
  {:else if !mostRead.page.items.length}<StateMessage kind="empty" title={text($locale,'Әзірге тайтл жоқ','Тайтлов пока нет')}/>
  {:else}<CoverShelf items={mostRead.page.items} label={text($locale,'Ең көп оқылатын','Самое читаемое')} eager/>{/if}
 </section>
 <section aria-labelledby="continue-heading">
  <div class="section-title"><h2 id="continue-heading">{text($locale,'Оқуды жалғастыру','Продолжить чтение')}</h2><a class="section-link" href="/library" aria-label={text($locale,'Кітапхана','Библиотека')}><Icon name="bookmark" size={18}/></a></div>
  {#if continuation}{@render continuation()}{/if}
 </section>
 <section aria-labelledby="updates-heading">
  <div class="section-title"><h2 id="updates-heading">{text($locale,'Соңғы жаңартулар','Последние обновления')}</h2><a class="section-link" href="/updates" aria-label={text($locale,'Барлық жаңартулар','Все обновления')}><Icon name="right" size={18}/></a></div>
  {#if updates.available}<Updates items={updates.page.items.slice(0,12)} compact/>{:else}<StateMessage kind="error" title={text($locale,'Жаңартулар жүктелмеді','Не удалось загрузить обновления')}/>{/if}
 </section>
 <section id="tops" aria-labelledby="tops-heading">
  <h2 id="tops-heading" class="sr-only">{text($locale,'Түрлері бойынша үздіктер','Топы по типам')}</h2>
  <nav class="shelf-tabs" aria-label={text($locale,'Түрі бойынша топ','Топ по типу')}>{#each tabs as tab}<a href={'/?top='+tab.value+'#tops'} aria-current={topKind===tab.value?'page':undefined}>{tab.label}</a>{/each}</nav>
  {#if !top.available}<StateMessage kind="error" title={text($locale,'Рейтинг жүктелмеді','Рейтинг не загрузился')}>{#snippet actions()}<a class="retry-link" href={'/?top='+topKind+'#tops'}>{text($locale,'Қайталау','Повторить')}</a>{/snippet}</StateMessage>
  {:else if !top.page.items.length}<StateMessage kind="empty" title={text($locale,'Бұл бөлімде әзірге тайтл жоқ','В этом разделе пока нет тайтлов')}/>
  {:else}<CoverShelf items={top.page.items} label={tabs.find(tab=>tab.value===topKind)?.label||tabs[0].label}/>{/if}
 </section>
</div>
<style>
 .discovery-home{display:grid;gap:28px;min-width:0}.discovery-home>section{min-width:0}.section-title{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:8px}h1,h2{font-size:19px;font-weight:650;line-height:1.35;overflow-wrap:anywhere}.section-link{border:1px solid var(--line);border-radius:8px;background:var(--panel);padding:0 10px;display:flex;align-items:center;justify-content:center;min-height:44px;min-width:44px;color:var(--muted);flex-shrink:0}.section-link:hover{color:var(--text)}.shelf-tabs{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));max-width:100%;gap:2px;padding:3px;margin-bottom:12px;border:1px solid var(--line);border-radius:12px}.shelf-tabs a{min-height:44px;display:flex;align-items:center;justify-content:center;padding:0 4px;border-radius:8px;font-size:11px;font-weight:550;white-space:nowrap;color:var(--muted)}.shelf-tabs a[aria-current]{background:var(--raised);color:var(--text)}.shelf-tabs a:hover{color:var(--text)}.retry-link{min-height:44px;display:inline-flex;align-items:center;color:var(--accent)}#tops{scroll-margin-top:80px}@media(min-width:768px){.discovery-home{gap:36px}h1,h2{font-size:24px}.shelf-tabs{width:max-content}.shelf-tabs a{font-size:14px;padding-inline:12px}}
</style>
