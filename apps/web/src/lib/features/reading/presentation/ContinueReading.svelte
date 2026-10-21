<script lang="ts">
 import {onMount,type Snippet} from 'svelte';
 import type {ReadingGateway,ProgressStorage,ReadingChanges} from '../application/progress';
 import {loadContinuation,type Continuation} from '../application/continuation';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {readingGateway,progressStorage,readingChanges,children}:{children?:Snippet<[boolean,Snippet]>;readingChanges?:ReadingChanges;readingGateway:ReadingGateway;progressStorage:ProgressStorage}=$props();
 const {locale}=useLocale();
 let rows=$state<Continuation[]>([]);
 let status=$state<'loading'|'ready'|'error'>('loading');
 let valid=true,generation=0;
 async function load(){const current=++generation;rows=[];status='loading';try{const loaded=await loadContinuation(readingGateway,progressStorage);if(valid&&current===generation){rows=loaded;status='ready';}}catch{if(valid&&current===generation)status='error';}}
 onMount(()=>{
  const stop=progressStorage.watchAccountChange(()=>{rows=[];void load();});
  const stopImports=readingChanges?.subscribe(()=>{void load();});void load();
  return()=>{valid=false;generation++;stop();stopImports?.();};
 });
</script>
{#snippet content()}
{#if status==='loading'}<div class="loading" role="status" aria-label={text($locale,'Оқу орны жүктелуде','Загружается продолжение чтения')}><span aria-hidden="true"></span></div>
{:else if rows.length}<div class="continuation-grid">{#each rows as item (item.titleId)}<a class="continue-card" href={item.availability==='available'||item.availability==='last-known'?'/read/'+encodeURIComponent(item.releaseId):'/titles/'+encodeURIComponent(item.slug)}>{#if item.coverURL}<img src={item.coverURL} alt="" width="112" height="160" loading="lazy" decoding="async"/>{:else}<span class="cover-fallback"><Icon name="bookmark"/></span>{/if}<span class="continue-name">{item.name}<small class="continue-meta">{item.availability==='withdrawn'?text($locale,'Жарияланым алынды','Публикация снята'):item.availability==='updated'?text($locale,'Жаңартылды · тарауды таңда','Обновлено · выбери главу'):text($locale,(item.chapterLabel?item.chapterLabel+'-тарау · ':'')+item.pagePosition+'-бет',(item.chapterLabel?'Глава '+item.chapterLabel+' · ':'')+'стр. '+item.pagePosition)}{item.teamName?' · '+item.teamName:''}{item.availability==='last-known'?' · '+text($locale,'құрылғыдағы орын','место на устройстве'):''}</small></span><span class="continue-action"><Icon name="right"/></span></a>{/each}</div>
{:else}<div class="continue-empty"><p>{status==='error'?text($locale,'Оқу орны жүктелмеді','Не удалось загрузить место чтения'):text($locale,'Әзірге оқу орны жоқ','Пока нечего продолжать')}</p><a class="button subtle" href={status==='error'?'/library':'/catalog'}>{status==='error'?text($locale,'Кітапхана','Библиотека'):'Каталог'}<Icon name="right" size={16}/></a></div>{/if}
{/snippet}
{#if children}{@render children(rows.length>0,content)}{:else}{@render content()}{/if}
<style>
 .continue-meta{display:block;color:var(--muted);font-size:12px;font-weight:400;margin-top:4px}
 .continuation-grid{display:grid;gap:12px}.continue-card{display:flex;align-items:center;gap:12px;min-width:0;background:var(--panel);padding:8px;border-radius:8px;border:1px solid var(--line)}img,.cover-fallback{width:48px;height:68px;object-fit:cover;border-radius:4px;flex-shrink:0}.cover-fallback{display:grid;place-items:center;background:var(--raised);color:var(--muted)}.continue-name{flex:1;min-width:0;font-size:14px;font-weight:550;overflow-wrap:anywhere;display:-webkit-box;-webkit-line-clamp:2;line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}.continue-action{display:grid;place-items:center;flex:0 0 36px;height:36px;border-radius:50%;background:var(--accent);color:var(--accent-ink)}.continue-card:hover{border-color:var(--muted)}.loading{height:84px;border-radius:8px;background:var(--panel);border:1px solid var(--line)}.loading span{display:block;width:48px;height:68px;margin:8px;background:var(--raised);border-radius:4px}.continue-empty{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;padding:16px;background:var(--panel);border:1px solid var(--line);border-radius:8px}.continue-empty p{font-size:14px;color:var(--muted)}@media(min-width:640px){.continuation-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}
</style>
