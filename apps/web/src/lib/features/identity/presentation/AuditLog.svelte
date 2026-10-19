<script lang="ts">
 import {onMount} from 'svelte';
 import {auditActorLabel,auditCoverValue,coverAuditCopy} from './audit-labels';
 import type {Account} from '../domain/account';
 import type {AuditEntry,ManagementGateway} from '../application/management';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 let {user,gateway}:{user:Account|null;gateway:ManagementGateway}=$props();
 const {locale}=useLocale();
 const coverCopy=$derived(coverAuditCopy($locale));
 const actions=$derived<Record<string,string>>({select_cover:coverCopy.select_cover,remove_cover:coverCopy.remove_cover,create:text($locale,'Құру','Создание'),update:text($locale,'Өзгерту','Изменение'),approve:text($locale,'Мақұлдау','Одобрение'),publish:text($locale,'Жариялау','Публикация'),reject:text($locale,'Түзетуге қайтару','Возврат на исправление'),withdraw:text($locale,'Жарияланымды алып тастау','Снятие публикации'),resolve_report:text($locale,'Шағым бойынша шешім','Решение по жалобе'),hide_comment:text($locale,'Пікірді жасыру','Скрытие комментария'),restore_comment:text($locale,'Пікірді қайта ашу','Восстановление комментария'),comment_visibility:text($locale,'Пікір көрінуін өзгерту','Изменение видимости комментария'),restore:text($locale,'Жарияланымды қалпына келтіру','Восстановление публикации'),manage_account:text($locale,'Аккаунтты басқару','Управление аккаунтом')});
 const entities=$derived<Record<string,string>>({title:text($locale,'Тайтл','Тайтл'),chapter:text($locale,'Тарау','Глава'),project:text($locale,'Жоба','Проект'),team:text($locale,'Команда','Команда'),release:text($locale,'Жарияланым','Публикация'),revision:text($locale,'Нұсқа','Версия'),community:text($locale,'Талқылау','Обсуждение'),user:text($locale,'Аккаунт','Аккаунт')});
 const fields=$derived<Record<string,string>>({cover_asset_id:coverCopy.field,role:text($locale,'Рөл','Роль'),blocked:text($locale,'Блокталған','Блокировка'),hidden:text($locale,'Жасырылған','Скрыт'),status:text($locale,'Күй','Статус'),revision_status:text($locale,'Нұсқа күйі','Статус версии'),release_status:text($locale,'Жарияланым күйі','Статус публикации')});
 const values=$derived<Record<string,string>>({reader:text($locale,'Оқырман','Читатель'),moderator:text($locale,'Модератор','Модератор'),admin:text($locale,'Әкімші','Администратор'),pending:text($locale,'Күтуде','Ожидает'),resolved:text($locale,'Шешілді','Решено'),dismissed:text($locale,'Негізсіз','Отклонено'),published:text($locale,'Жарияланған','Опубликовано'),withdrawn:text($locale,'Алып тасталған','Снято'),draft:text($locale,'Жоба','Черновик'),ready:text($locale,'Дайын','Готово'),processing:text($locale,'Өңделуде','Обработка'),rejected:text($locale,'Қайтарылды','Отклонено')});
 const stateValue=(value:unknown)=>typeof value==='boolean'?(value?text($locale,'Иә','Да'):text($locale,'Жоқ','Нет')):typeof value==='string'?(values[value]||value):value===undefined?'—':String(value);
 let items=$state<AuditEntry[]>([]),hasMore=$state(false),busy=$state(false),message=$state('');
 async function load(after=''){if(!user)return;busy=true;message='';try{const result=await gateway.audit(user.id,after);items=after?[...items,...result.items]:result.items;hasMore=result.hasMore;}catch(failure){message=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{if(user?.role==='admin')void load();});
</script>
<svelte:head><title>{text($locale,'Әрекеттер журналы','Журнал действий')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Әрекеттер журналы','Журнал действий')}/>
{#if user?.role!=='admin'}<p>{text($locale,'Әкімші рұқсаты қажет.','Нужны права администратора.')}</p>{:else}
{#if message}<p role="alert">{noticeText($locale,message)}</p>{/if}
{#each items as item}<article>
 <p><strong>{actions[item.action]||item.action}</strong> · <time datetime={item.createdAt}>{new Date(item.createdAt).toLocaleString($locale==='ru'?'ru-RU':'kk-KZ')}</time></p>
 <p>{text($locale,'Кім','Кто')}: {#if item.actor}<a class="auth-link" href={'/moderation/accounts?q='+encodeURIComponent(item.actor)}>{auditActorLabel(item.actorName,item.actor,$locale)}</a>{:else}{auditActorLabel(item.actorName,item.actor,$locale)}{/if}</p>
 <p>{entities[item.entityType]||item.entityType}: {#if item.targetPath}<a class="auth-link" href={item.targetPath}>{item.targetName||text($locale,'Нысанды ашу','Открыть объект')}</a>{:else}{item.targetName||text($locale,'Нысан қолжетімсіз','Объект недоступен')}{/if}</p>
 {#if Object.keys(item.before).length||Object.keys(item.after).length}<dl>{#each [...new Set([...Object.keys(item.before),...Object.keys(item.after)])] as key}<dt>{fields[key]||key}</dt><dd>{key==='cover_asset_id'?auditCoverValue(item.before[key],$locale):stateValue(item.before[key])} → {key==='cover_asset_id'?auditCoverValue(item.after[key],$locale):stateValue(item.after[key])}</dd>{/each}</dl>{:else}<p class="muted">{text($locale,'Бұл жазбада бұрынғы және жаңа күй сақталмаған.','В этой записи прежнее и новое состояние не сохранены.')}</p>{/if}
 {#if item.reason}<p>{text($locale,'Себеп','Причина')}: {item.reason}</p>{/if}
 <details><summary>{text($locale,'Техникалық деректер','Технические данные')}</summary><p>{text($locale,'Әрекет жасаған аккаунт ID','ID исполнителя')}: {item.actor||'—'}</p><p>{text($locale,'Нысан ID','ID объекта')}: {item.entityId}</p><p>ID: {item.id}</p>{#if 'cover_asset_id' in item.before||'cover_asset_id' in item.after}<p>{coverCopy.field} ID: {item.before.cover_asset_id||'—'} → {item.after.cover_asset_id||'—'}</p>{/if}</details>
 </article>{:else}{#if !busy}<p>{text($locale,'Жазбалар жоқ','Записей нет')}</p>{/if}{/each}
{#if hasMore}<button class="button subtle" disabled={busy} onclick={()=>load(items[items.length-1].id)}>{text($locale,'Тағы жүктеу','Загрузить ещё')}</button>{/if}{/if}
<style>summary{display:flex;align-items:center;min-height:44px;cursor:pointer}dl{display:grid;grid-template-columns:minmax(90px,1fr) 2fr;gap:8px;font-size:14px;margin:12px 0}dd{margin:0;overflow-wrap:anywhere}article{max-width:800px;padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);margin-block:12px;min-width:0}p{overflow-wrap:anywhere;white-space:pre-wrap;margin-bottom:8px;font-size:14px}</style>
