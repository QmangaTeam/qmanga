<script lang="ts">
 import {onMount} from 'svelte';
 import type {Notification,NotificationsGateway} from '../application/notifications';
 let {gateway}:{gateway:NotificationsGateway}=$props();
 let account=$state<string|null>(null),items=$state<Notification[]>([]),busy=$state(true),more=$state(false),message=$state('');
 async function load(after=''){busy=true;try{account=await gateway.account();if(account){const result=await gateway.list(account,after);items=after?[...items,...result.items]:result.items;more=result.hasMore;}}catch(failure){message=failure instanceof Error?failure.message:'Хабарлар жүктелмеді';}finally{busy=false;}}
 onMount(()=>{void load();});
 async function read(item:Notification){if(!account)return;try{await gateway.read(account,item.id);item.read=true;items=[...items];}catch{message='Хабар белгіленбеді';}}
</script>
<svelte:head><title>Хабарлар — Qmanga</title><meta name="robots" content="noindex"/></svelte:head><h1>Хабарлар</h1><p class="muted">Кітапханадағы жаңа тараулар мен пікіріңізге жауаптар.</p>
{#if message}<p class="form-message" role="status">{message}</p>{/if}{#if !account&&!busy}<p>Хабарларды көру үшін <a class="auth-link" href="/login">кіріңіз</a>.</p>{/if}
{#each items as item}<article class:unread={!item.read}><h2>{item.title}</h2><p>{item.body}</p><time datetime={item.createdAt}>{new Date(item.createdAt).toLocaleString('kk-KZ')}</time><div>{#if item.releaseId}<a class="button primary" href={`/read/${encodeURIComponent(item.releaseId)}`} onclick={()=>read(item)}>Оқу</a>{:else if item.titleSlug}<a class="button subtle" href={`/titles/${encodeURIComponent(item.titleSlug)}`} onclick={()=>read(item)}>Тайтлды ашу</a>{/if}{#if !item.read}<button class="button subtle" onclick={()=>read(item)}>Оқылды деп белгілеу</button>{/if}</div></article>{:else}{#if account&&!busy}<p class="empty-state">Әзірге хабар жоқ.</p>{/if}{/each}{#if more}<button class="button" onclick={()=>load(items[items.length-1].id)} disabled={busy}>Тағы жүктеу</button>{/if}{#if busy}<p role="status">Жүктелуде…</p>{/if}
<style>h1{font-size:26px}article{border-bottom:1px solid var(--line);padding:20px 12px}article.unread{border-left:3px solid var(--accent)}h2{font-size:17px}p{margin:8px 0}time{font-size:12px;color:var(--muted)}article>div{display:flex;flex-wrap:wrap;gap:10px;margin-top:12px}</style>
