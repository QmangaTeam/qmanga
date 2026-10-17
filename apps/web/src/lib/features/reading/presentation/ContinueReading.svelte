<script lang="ts">
 import {onMount} from 'svelte';
 import type {ReadingGateway,ProgressStorage} from '../application/progress';
 import {loadContinuation} from '../application/continuation';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {readingGateway,progressStorage,covers=[]}:{readingGateway:ReadingGateway;progressStorage:ProgressStorage;covers?:{id:string;coverURL:string}[]}=$props();
 const {locale}=useLocale();
 let rows=$state<{titleId:string;name:string;slug:string;releaseId:string}[]>([]);
 let valid=true;
 onMount(()=>{
  const stop=progressStorage.watchAccountChange(()=>{valid=false;rows=[];});
  void(async()=>{try{
   const loaded=await loadContinuation(readingGateway,progressStorage);
   if(valid)rows=loaded;
  }catch{rows=[];}})();
  return()=>{valid=false;stop();};
 });
</script>
{#if rows.length}<section class="continue-reading"><div class="continuation-heading"><h2>{text($locale,'Оқуды жалғастыр','Продолжить чтение')}</h2><a href="/library" class="library-link" aria-label={text($locale,'Кітапхана','Библиотека')}><Icon name="bookmark"/></a></div><div class="continuation-grid">{#each rows as item (item.titleId)}{@const coverURL=covers.find(cover=>cover.id===item.titleId)?.coverURL}<a class="continue-card" href={'/read/'+encodeURIComponent(item.releaseId)}>{#if coverURL}<img src={coverURL} alt="" width="112" height="160" loading="lazy" decoding="async"/>{:else}<span class="cover-fallback"><Icon name="bookmark"/></span>{/if}<span class="continue-name">{item.name}</span><span class="continue-action"><Icon name="right"/></span></a>{/each}</div></section>{/if}
<style>
 .continue-reading{min-width:0}.continuation-heading{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:8px}h2{font-size:19px;font-weight:650;line-height:1.35}.library-link{display:grid;place-items:center;min-width:44px;min-height:44px;background:var(--panel);border:1px solid var(--line);border-radius:8px}.continuation-grid{display:grid;gap:12px}.continue-card{display:flex;align-items:center;gap:12px;min-width:0;background:var(--panel);padding:8px;border-radius:8px;border:1px solid var(--line)}img,.cover-fallback{width:48px;height:68px;object-fit:cover;border-radius:4px;flex-shrink:0}.cover-fallback{display:grid;place-items:center;background:var(--raised);color:var(--muted)}.continue-name{flex:1;min-width:0;font-size:14px;font-weight:550;overflow-wrap:anywhere;display:-webkit-box;-webkit-line-clamp:2;line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}.continue-action{display:grid;place-items:center;flex:0 0 36px;height:36px;border-radius:50%;background:var(--accent);color:var(--accent-ink)}.continue-card:hover{border-color:var(--muted)}@media(min-width:640px){.continuation-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(min-width:768px){h2{font-size:24px}}
</style>
