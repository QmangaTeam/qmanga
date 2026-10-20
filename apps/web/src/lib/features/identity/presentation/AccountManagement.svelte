<script lang="ts">
 import UnsavedChanges from '$lib/shared/presentation/ui/UnsavedChanges.svelte';
 import {fieldValidation} from '$lib/shared/presentation/form-validation';
 import {onMount} from 'svelte';
 import {page} from '$app/state';
 import type {Account} from '../domain/account';
 import type {ManagedAccount,ManagementGateway} from '../application/management';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 let {user,gateway}:{user:Account|null;gateway:ManagementGateway}=$props();
 const {locale}=useLocale();
 const roles=$derived<Record<string,string>>({reader:text($locale,'Оқырман','Читатель'),moderator:text($locale,'Модератор','Модератор'),admin:text($locale,'Әкімші','Администратор')});
 let query=$state(page.url.searchParams.get('q')||''),items=$state<ManagedAccount[]>([]),hasMore=$state(false),busy=$state(false),message=$state(''),selected=$state<ManagedAccount|null>(null),reason=$state(''),saved=$state(false);
 let appliedQuery='';
 let guard:UnsavedChanges;
 let original=$state('');
 const dirty=$derived(!!selected&&(JSON.stringify(selected)!==original||reason!==''));
 function choose(item:ManagedAccount){if(!guard.confirmDiscard())return;selected={...item};original=JSON.stringify(selected);reason='';saved=false;}
 function cancel(){if(guard.confirmDiscard()){selected=null;reason='';}}
 const allowed=$derived(user?.role==='admin'||user?.role==='moderator');
 async function load(after=''){
  if(!user)return;
  if(!after&&selected){if(!guard.confirmDiscard())return;selected=null;reason='';}
  const requestedQuery=after?appliedQuery:query;
  busy=true;message='';
  try{const result=await gateway.accounts(user.id,requestedQuery,after);items=after?[...items,...result.items]:result.items;hasMore=result.hasMore;if(!after)appliedQuery=requestedQuery;}
  catch(failure){message=failureCode(failure);}finally{busy=false;}
 }
 async function save(){if(!user||!selected)return;busy=true;message='';saved=false;try{await gateway.change(user.id,selected,reason);selected=null;reason='';await load();saved=true;}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{if(allowed)void load();});
</script>
<UnsavedChanges bind:this={guard} {dirty}/>
<svelte:head><title>{text($locale,'Аккаунттарды басқару','Управление аккаунтами')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Аккаунттарды басқару','Управление аккаунтами')} description={text($locale,'Рөлдер мен блоктау. Әр өзгеріс журналға жазылады.','Роли и блокировки. Каждое изменение записывается в журнал.')}/>
{#if !allowed}<p>{text($locale,'Модератор рұқсаты қажет.','Нужны права модератора.')}</p>{:else}
<form class="search" onsubmit={event=>{event.preventDefault();void load();}}><label>{text($locale,'Аты, ID немесе толық пошта','Имя, ID или точная почта')}<input bind:value={query} maxlength="160"/></label><button class="button subtle" disabled={busy}>{text($locale,'Іздеу','Найти')}</button></form>
{#if message}<p role="alert" class="form-message">{noticeText($locale,message)}</p>{/if}{#if saved}<p role="status">{text($locale,'Өзгеріс сақталды','Изменение сохранено')}</p>{/if}
{#each items as item}<article><div><strong>{item.name||text($locale,'Оқырман','Читатель')}</strong><small>{item.id} · {roles[item.role]} · {item.blocked?text($locale,'Блокталған','Заблокирован'):text($locale,'Белсенді','Активен')}</small></div><button class="button subtle" disabled={busy||item.id===user?.id||(user?.role!=='admin'&&item.role!=='reader')} onclick={()=>choose(item)}>{text($locale,'Өзгерту','Изменить')}</button></article>{#if selected?.id===item.id}<form class="edit" use:fieldValidation={{locale:$locale}} onsubmit={event=>{event.preventDefault();void save();}}><h2>{selected.name||selected.id}</h2><label>{text($locale,'Рөл','Роль')}<select bind:value={selected.role} disabled={user?.role!=='admin'}><option value="reader">{text($locale,'Оқырман','Читатель')}</option><option value="moderator">{text($locale,'Модератор','Модератор')}</option><option value="admin">{text($locale,'Әкімші','Администратор')}</option></select></label><label class="check"><input type="checkbox" bind:checked={selected.blocked}/>{text($locale,'Аккаунтты блоктау','Заблокировать аккаунт')}</label><label>{text($locale,'Себеп','Причина')}<textarea bind:value={reason} required minlength="8" maxlength="2000" rows="3"></textarea></label><p class="muted">{text($locale,'Аккаунттың барлық сессиясы жабылады. Өз аккаунтыңды өзгертуге болмайды.','Все сессии аккаунта завершатся. Нельзя изменить собственный аккаунт.')}</p><div class="actions"><button class="button primary" disabled={busy}>{text($locale,'Өзгерісті растау','Подтвердить изменение')}</button><button type="button" class="button subtle" onclick={cancel}>{text($locale,'Бас тарту','Отмена')}</button></div></form>{/if}
{:else}{#if !busy}<p>{text($locale,'Табылмады','Ничего не найдено')}</p>{/if}{/each}
{#if hasMore}<button class="button subtle" disabled={busy} onclick={()=>load(items[items.length-1].id)}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}
{#if user?.role==='admin'}<a class="auth-link" href="/admin/audit">{text($locale,'Әрекеттер журналы','Журнал действий')}</a>{/if}{/if}
<style>form{max-width:680px;display:grid;gap:16px;margin:20px 0}label{display:grid;gap:8px}article{max-width:800px;display:flex;align-items:center;justify-content:space-between;gap:12px;padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);margin-block:12px;min-width:0}small{display:block;overflow-wrap:anywhere;color:var(--muted);margin-top:6px}.actions{display:flex;flex-wrap:wrap;gap:8px}.check{display:flex;align-items:center;min-height:44px}.check input{height:20px;min-height:20px;width:20px}.edit{padding:20px 0;border-block:1px solid var(--line)}.search button{justify-self:start}article>div{min-width:0}.edit{scroll-margin-top:100px}</style>
