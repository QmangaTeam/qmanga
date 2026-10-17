<script lang="ts">
 import {onMount} from 'svelte';
 import type {Account} from '../domain/account';
 import type {AuditEntry,ManagementGateway} from '../application/management';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 let {user,gateway}:{user:Account|null;gateway:ManagementGateway}=$props();
 const {locale}=useLocale();
 const actions=$derived<Record<string,string>>({create:text($locale,'Құру','Создание'),update:text($locale,'Өзгерту','Изменение'),approve:text($locale,'Мақұлдау','Одобрение'),publish:text($locale,'Жариялау','Публикация'),reject:text($locale,'Түзетуге қайтару','Возврат на исправление'),withdraw:text($locale,'Жарияланымды алып тастау','Снятие публикации'),resolve_report:text($locale,'Шағым бойынша шешім','Решение по жалобе'),hide_comment:text($locale,'Пікірді жасыру','Скрытие комментария'),manage_account:text($locale,'Аккаунтты басқару','Управление аккаунтом')});
 const entities=$derived<Record<string,string>>({title:text($locale,'Тайтл','Тайтл'),chapter:text($locale,'Тарау','Глава'),project:text($locale,'Жоба','Проект'),team:text($locale,'Команда','Команда'),release:text($locale,'Жарияланым','Публикация'),revision:text($locale,'Нұсқа','Версия'),community:text($locale,'Талқылау','Обсуждение'),user:text($locale,'Аккаунт','Аккаунт')});
 let items=$state<AuditEntry[]>([]),hasMore=$state(false),busy=$state(false),message=$state('');
 async function load(after=''){if(!user)return;busy=true;message='';try{const result=await gateway.audit(user.id,after);items=after?[...items,...result.items]:result.items;hasMore=result.hasMore;}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{if(user?.role==='admin')void load();});
</script>
<svelte:head><title>{text($locale,'Әрекеттер журналы','Журнал действий')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Әрекеттер журналы','Журнал действий')}/>
{#if user?.role!=='admin'}<p>{text($locale,'Әкімші рұқсаты қажет.','Нужны права администратора.')}</p>{:else}
{#if message}<p role="alert">{noticeText($locale,message)}</p>{/if}
{#each items as item}<article><p><strong>{actions[item.action]||item.action}</strong> · {new Date(item.createdAt).toLocaleString($locale==='ru'?'ru-RU':'kk-KZ')}</p><p>{text($locale,'Кім','Кто')}: {item.actor||'—'}</p><p>{entities[item.entityType]||item.entityType}: {item.entityId}</p>{#if item.reason}<p>{item.reason}</p>{/if}</article>{:else}{#if !busy}<p>{text($locale,'Жазбалар жоқ','Записей нет')}</p>{/if}{/each}
{#if hasMore}<button class="button subtle" disabled={busy} onclick={()=>load(items[items.length-1].id)}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}{/if}
<style>article{max-width:800px;padding:20px 0;border-bottom:1px solid var(--line)}p{overflow-wrap:anywhere;white-space:pre-wrap;margin-bottom:8px;font-size:14px}</style>
