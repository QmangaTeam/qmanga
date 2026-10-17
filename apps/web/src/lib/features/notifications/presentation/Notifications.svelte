<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {onMount} from 'svelte';
 import type {Notification,NotificationsGateway} from '../application/notifications';
 let {gateway}:{gateway:NotificationsGateway}=$props();
 let account=$state<string|null>(null),items=$state<Notification[]>([]),busy=$state(true),more=$state(false),message=$state('');
 async function load(after=''){busy=true;try{account=await gateway.account();if(account){const result=await gateway.list(account,after);items=after?[...items,...result.items]:result.items;more=result.hasMore;}}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{void load();});
 async function read(item:Notification){if(!account)return;try{await gateway.read(account,item.id);item.read=true;items=[...items];}catch{message="notification_failed";}}
</script>
<svelte:head><title>{copy($locale,"Хабарлар — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head><h1>{copy($locale,"Хабарлар")}</h1><p class="muted">{copy($locale,"Кітапханадағы жаңа тараулар мен пікіріңізге жауаптар.")}</p>
{#if message}<p class="form-message" role="status">{noticeText($locale,message)}</p>{/if}{#if !account&&!busy}<p>{copy($locale,"Хабарларды көру үшін ")}<a class="auth-link" href="/login">{copy($locale,"кіріңіз")}</a>.</p>{/if}
{#each items as item}<article class:unread={!item.read}><h2>{item.kind==='new_release'?text($locale,'Жаңа тарау','Новая глава'):item.kind==='comment_reply'?text($locale,'Пікірге жауап','Ответ на комментарий'):copy($locale,'Хабарлар')}</h2><p lang="kk">{item.body}</p><time datetime={item.createdAt}>{new Date(item.createdAt).toLocaleString($locale==='ru'?'ru-RU':'kk-KZ')}</time><div>{#if item.releaseId}<a class="button primary" href={`/read/${encodeURIComponent(item.releaseId)}`} onclick={()=>read(item)}>{copy($locale,"Оқу")}</a>{:else if item.titleSlug}<a class="button subtle" href={`/titles/${encodeURIComponent(item.titleSlug)}`} onclick={()=>read(item)}>{copy($locale,"Тайтлды ашу")}</a>{/if}{#if !item.read}<button class="button subtle" onclick={()=>read(item)}>{copy($locale,"Оқылды деп белгілеу")}</button>{/if}</div></article>{:else}{#if account&&!busy}<p class="empty-state">{copy($locale,"Әзірге хабар жоқ.")}</p>{/if}{/each}{#if more}<button class="button" onclick={()=>load(items[items.length-1].id)} disabled={busy}>{copy($locale,"Тағы жүктеу")}</button>{/if}{#if busy}<p role="status">{copy($locale,"Жүктелуде…")}</p>{/if}
<style>h1{font-size:26px}article{border-bottom:1px solid var(--line);padding:20px 12px}article.unread{border-left:3px solid var(--accent)}h2{font-size:17px}p{margin:8px 0}time{font-size:12px;color:var(--muted)}article>div{display:flex;flex-wrap:wrap;gap:10px;margin-top:12px}</style>
