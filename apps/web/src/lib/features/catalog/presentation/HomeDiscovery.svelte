<script lang="ts">
 import type {Snippet} from 'svelte';
 import type {ResultPage,Summary,ReleaseUpdate} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 import CoverShelf from './CoverShelf.svelte';
 import Updates from './Updates.svelte';
 let {catalogue,updates,picks,shelf='updated',continuation}:{catalogue:{available:boolean;page:ResultPage<Summary>};updates:{available:boolean;page:ResultPage<ReleaseUpdate>};picks:{available:boolean;page:ResultPage<Summary>};shelf?:string;continuation?:Snippet}=$props();
 const {locale}=useLocale();
 const tabs=$derived([{value:'updated',label:text($locale,'Жаңарған','Обновлённые')},{value:'new',label:text($locale,'Жаңа','Новинки')},{value:'completed',label:text($locale,'Аяқталған','Завершённые')}]);
</script>
<div class="discovery-home">
 <section aria-labelledby="browse-heading">
  <div class="section-title"><h1 id="browse-heading">{text($locale,'Не оқимыз?','Что почитать')}</h1><a class="section-link" href="/catalog">{text($locale,'Каталог','Каталог')}<Icon name="right" size={16}/></a></div>
  <nav class="shelf-tabs" aria-label={text($locale,'Тайтлдарды реттеу','Порядок тайтлов')}>{#each tabs as tab}<a href={tab.value==='updated'?'/':'/?shelf='+tab.value} aria-current={shelf===tab.value?'page':undefined}>{tab.label}</a>{/each}</nav>
  {#if !catalogue.available}<StateMessage kind="error" title={text($locale,'Каталог жүктелмеді','Каталог не загрузился')}>{#snippet actions()}<a class="retry-link" href={shelf==='updated'?'/':'/?shelf='+shelf}>{text($locale,'Қайталау','Повторить')}</a>{/snippet}</StateMessage>
  {:else if !catalogue.page.items.length}<StateMessage kind="empty" title={text($locale,'Әзірге тайтл жоқ','Тайтлов пока нет')}/>
  {:else}<CoverShelf items={catalogue.page.items.slice(0,16)} label={text($locale,'Оқуға арналған тайтлдар','Тайтлы для чтения')}/>{/if}
 </section>
 {#if continuation}{@render continuation()}{/if}
 {#if picks.available&&picks.page.items.length}<section aria-labelledby="picks-heading"><div class="section-title"><h2 id="picks-heading">{text($locale,'Редакция таңдауы','Подборка редакции')}</h2></div><CoverShelf items={picks.page.items.slice(0,16)} label={text($locale,'Редакция таңдауы','Подборка редакции')}/></section>{/if}
 <section aria-labelledby="updates-heading">
  <div class="section-title"><h2 id="updates-heading">{text($locale,'Жаңа тараулар','Новые главы')}</h2><a class="section-link" href="/updates">{text($locale,'Барлығы','Все')}<Icon name="right" size={16}/></a></div>
  {#if updates.available}<Updates items={updates.page.items.slice(0,12)} compact/>{:else}<StateMessage kind="error" title={text($locale,'Жаңартулар жүктелмеді','Не удалось загрузить обновления')}/>{/if}
 </section>
</div>
<style>
 .discovery-home{display:grid;gap:28px;min-width:0}.discovery-home>section{min-width:0}.section-title{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:8px}h1,h2{font-size:19px;font-weight:650;line-height:1.35;overflow-wrap:anywhere}.section-link{border:1px solid var(--line);border-radius:8px;background:var(--panel);padding:0 10px;display:flex;align-items:center;gap:4px;min-height:44px;min-width:44px;color:var(--muted);font-size:12px;flex-shrink:0}.section-link:hover{color:var(--text)}.shelf-tabs{display:flex;width:max-content;max-width:100%;overflow-x:auto;gap:2px;padding:3px;margin-bottom:12px;border:1px solid var(--line);border-radius:12px}.shelf-tabs a{min-height:44px;display:flex;align-items:center;justify-content:center;padding:0 8px;border-radius:8px;font-size:12px;font-weight:550;white-space:nowrap;color:var(--muted)}.shelf-tabs a[aria-current]{background:var(--raised);color:var(--text)}.shelf-tabs a:hover{color:var(--text)}.retry-link{min-height:44px;display:inline-flex;align-items:center;color:var(--accent)}@media(min-width:768px){.discovery-home{gap:36px}h1,h2{font-size:24px}.shelf-tabs a{font-size:14px;padding-inline:12px}.section-link{font-size:14px}}
</style>
