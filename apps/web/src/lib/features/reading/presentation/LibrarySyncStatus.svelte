<script lang="ts">
 import {onMount} from 'svelte';
 import type {LibrarySync,LibraryLocal} from '../application/library-sync';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 let {sync,account,onresolved}:{sync:LibrarySync;account:string;onresolved?:()=>void}=$props();
 const {locale}=useLocale();
 let rows=$state<LibraryLocal[]>([]),busy=$state(false),error=$state(''),active=true;
 async function refresh(){try{const next=await sync.rows(account);if(active)rows=next.filter(row=>row.pending);}catch(failure){if(active)error=failureCode(failure);}}
 async function flush(){if(busy||!active)return;busy=true;try{const pending=(await sync.rows(account)).some(row=>row.pending);await sync.flush(account);await refresh();if(pending)onresolved?.();}catch(failure){error=failureCode(failure);}finally{busy=false;}}
 async function resolve(row:LibraryLocal,useLocal:boolean){busy=true;try{await sync.resolve(account,row.selection.titleId,useLocal);await refresh();onresolved?.();}catch(failure){error=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{const stop=sync.subscribe(()=>{void refresh();});void flush();const online=()=>{void flush();};window.addEventListener('online',online);return()=>{active=false;stop();window.removeEventListener('online',online);};});
</script>
{#if error}<p role="alert">{noticeText($locale,error)}</p>{/if}
{#if rows.length}<section class="sync" aria-label={text($locale,'Кітапхананы синхрондау','Синхронизация библиотеки')}>
<p role="status">{text($locale,'Құрылғыда сақталған өзгерістер:','Сохранённые на устройстве изменения:')} {rows.length}. {text($locale,'Бұлтқа әлі жіберілмеген.','Ещё не отправлены в облако.')}</p>
{#each rows as row}<div class="pending"><strong>{row.entry?.name||row.selection.titleId}</strong>
{#if row.conflict}<p>{text($locale,'Басқа құрылғыдағы таңдау өзгеше. Қайсысын қалдырамыз?','На другом устройстве другой выбор. Какой оставить?')}</p><div class="actions"><button class="button primary" disabled={busy} onclick={()=>resolve(row,true)}>{text($locale,'Осы құрылғыдағыны','С этого устройства')}</button><button class="button subtle" disabled={busy} onclick={()=>resolve(row,false)}>{text($locale,'Бұлттағыны','Из облака')}</button></div>
{:else if row.error&&!['network','unavailable'].includes(row.error)}<p>{noticeText($locale,row.error)}</p><button class="button subtle" disabled={busy} onclick={()=>resolve(row,false)}>{text($locale,'Жергілікті өзгерісті алып тастау','Отменить локальное изменение')}</button>{/if}</div>{/each}
<button class="button subtle" disabled={busy} onclick={flush}>{text($locale,'Қайта синхрондау','Повторить синхронизацию')}</button>
</section>{/if}
<style>.sync{padding:20px 0;border-block:1px solid var(--line)}.pending{margin:16px 0}p{font-size:14px;overflow-wrap:anywhere;margin:8px 0}.actions{display:flex;flex-wrap:wrap;gap:8px}</style>
