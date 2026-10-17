<script lang="ts">
 import {onMount} from 'svelte';
 import type {CommunityGateway} from '../application/community';
 import type {PersonalReport} from '../application/rights';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {gateway}:{gateway:CommunityGateway}=$props();
 const {locale}=useLocale();
 let items=$state<PersonalReport[]>([]),more=$state(false),busy=$state(true),message=$state(''),account=$state<string|null>(null);
 async function load(after=''){busy=true;message='';try{account=await gateway.account();if(!account)return;const result=await gateway.myReports(account,after);items=after?[...items,...result.items]:result.items;more=result.hasMore;}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{void load();});
</script>
<svelte:head><title>{text($locale,'Менің өтініштерім','Мои обращения')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Менің өтініштерім','Мои обращения')}/>
{#if message}<StateMessage kind="error" title={noticeText($locale,message)}/>{/if}
{#if !busy&&!account}<a class="auth-link" href="/login?returnTo=%2Freports">{text($locale,'Кіру','Войти')}</a>{/if}
{#each items as item}<article><p class="muted">{new Date(item.createdAt).toLocaleDateString($locale==='ru'?'ru-RU':'kk-KZ')} · {item.status==='pending'?text($locale,'Қаралуда','На рассмотрении'):item.status==='resolved'?text($locale,'Шешілді','Решено'):text($locale,'Қабылданбады','Отклонено')}</p><p>{item.reason}</p>{#if item.resolution}<h2>{text($locale,'Модератор жауабы','Ответ модератора')}</h2><p>{item.resolution}</p>{/if}</article>{:else}{#if !busy&&account&&!message}<StateMessage kind="empty" title={text($locale,'Өтініштер әлі жоқ','Обращений пока нет')}/>{/if}{/each}
{#if more}<button class="button subtle" disabled={busy} onclick={()=>load(items[items.length-1].id)}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}
{#if busy}<p role="status">{text($locale,'Жүктелуде…','Загрузка…')}</p>{/if}
<style>article{max-width:720px;padding:20px 0;border-bottom:1px solid var(--line)}p{white-space:pre-wrap;overflow-wrap:anywhere;margin-bottom:12px}h2{font-size:16px;margin-bottom:8px}</style>
