<script lang="ts">
 import {queueDepth,queueReturn} from '$lib/shared/presentation/queue-context';
 import {tick} from 'svelte';
 import {page} from '$app/state';
 import QueueFilters from '$lib/shared/presentation/ui/QueueFilters.svelte';
 import {onMount} from 'svelte';import {useLocale,text} from '$lib/shared/presentation/locale';import {noticeText,failureCode} from '$lib/shared/presentation/copy';import type {TitleChangesGateway,TitleChange} from '../application/title-changes';
 let {gateway}:{gateway:TitleChangesGateway}=$props();const {locale}=useLocale();let actor=$state(''),items=$state<TitleChange[]>([]),hasMore=$state(false),loading=$state(false),error=$state('');let generation=0;let depth=$state(1);
 async function load(more=false){const run=more?generation:++generation;loading=true;error='';if(!more){items=[];actor='';}try{const account=more?actor:await gateway.currentAccount();if(!account)throw Error('unauthorized');const result=await gateway.queue(account,more?items.at(-1)?.id:'',page.url.searchParams.get('team')||'',page.url.searchParams.get('kind')||'');if(run!==generation)return;actor=account;items=more?[...items,...result.items]:result.items;hasMore=result.has_more;if(more)depth++;else depth=1;}catch(e){if(run===generation)error=failureCode(e);}finally{if(run===generation)loading=false;}}
 onMount(()=>{void (async()=>{await load();const wanted=queueDepth(page.url);for(let step=1;step<wanted&&hasMore&&!error;step++)await load(true);await tick();if(page.url.hash)document.getElementById(page.url.hash.slice(1))?.scrollIntoView();})();const stop=gateway.watchAccount(()=>void load());return()=>{generation++;stop();};});
</script>
<svelte:head><title>{text($locale,'Тайтл деректері','Данные тайтлов')} — Qmanga</title></svelte:head>
<h1>{text($locale,'Тайтл деректері','Данные тайтлов')}</h1><p class="muted">{text($locale,'Сипаттама, жанрлар, санаттар, мұқаба және баннер өзгерістері.','Изменения описания, жанров, категорий, обложки и баннера.')}</p>
<QueueFilters kinds={[{value:'metadata',kk:'Деректер',ru:'Данные'},{value:'artwork',kk:'Мұқаба немесе баннер',ru:'Обложка или баннер'}]}/>
{#if error}<p role="alert">{noticeText($locale,error)}</p><button class="button subtle" onclick={()=>load()}>{text($locale,'Қайталау','Повторить')}</button>{/if}
{#each items as item}<article class="panel" id={'queue-'+item.id}><h2><a href={'/moderation/title-changes/'+encodeURIComponent(item.id)+'?returnTo='+encodeURIComponent(queueReturn(page.url,depth,item.id))}>{item.title_name}</a></h2><p class="muted">{item.team_name} · {item.requested_name}</p><p>{item.reason}</p></article>{:else}{#if !loading&&!error}<p>{text($locale,'Қаралатын өтінім жоқ.','Нет заявок на рассмотрении.')}</p>{/if}{/each}
{#if loading}<p role="status">{text($locale,'Жүктелуде…','Загрузка…')}</p>{/if}{#if hasMore}<button class="button subtle" disabled={loading} onclick={()=>load(true)}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}
<style>h1{font-size:24px}article{margin-block:12px;overflow-wrap:anywhere}h2{font-size:18px;margin-top:0}article p{white-space:pre-wrap}</style>
