<script lang="ts">
 import {onMount} from 'svelte';
 import type {TeamsGateway} from '../application/teams';
 import type {Account} from '$lib/features/identity/domain/account';
 import type {Team} from '../domain/team';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 let {user,gateway}:{user:Account|null;gateway:TeamsGateway}=$props();const {locale}=useLocale();
 let teams=$state<Team[]>([]),loading=$state(true),busy=$state(false),more=$state(false),error=$state(''),selected=$state(''),reason=$state(''),notice=$state('');
 async function load(append=false){const batch=await gateway.pending(append?teams.length:0);teams=append?[...teams,...batch]:batch;more=batch.length===100;}
 async function run(action:()=>Promise<void>){busy=true;error='';try{await action();}catch(failure){error=failureCode(failure);}finally{busy=false;}}
 onMount(()=>{if(user?.role!=='admin'){loading=false;return;}void run(async()=>{try{await load();}finally{loading=false;}});});
</script>
<svelte:head><title>{text($locale,'Команда өтінімдері','Заявки команд')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Команда өтінімдері','Заявки команд')} description={text($locale,'Мақұлданған командалар жобалар сұрай алады. Қайтарған кезде себебін жаз.','Одобренные команды смогут запрашивать проекты. При возврате укажи причину.')}/>
{#if user?.role!=='admin'}<p>{text($locale,'Әкімші рұқсаты қажет.','Нужен доступ администратора.')}</p><a class="button primary" href="/login?returnTo=%2Fadmin">{text($locale,'Кіру','Войти')}</a>
{:else}
 {#if error}<p class="form-message" role="alert">{noticeText($locale,error)}</p>{/if}{#if notice}<p role="status">{notice}</p>{/if}
 {#if loading}<p role="status">{text($locale,'Өтінімдер жүктелуде…','Загружаем заявки…')}</p>{:else}
 {#each teams as team}<article><h2>{team.name}</h2><p>{team.description||text($locale,'Сипаттама берілмеген.','Описание не указано.')}</p><small>/teams/{team.slug}</small><div class="actions"><button class="button primary" disabled={busy} onclick={()=>run(async()=>{await gateway.approve(team.id);await load();notice=text($locale,'Команда мақұлданды.','Команда одобрена.');})}>{text($locale,'Мақұлдау','Одобрить')}</button><button class="button subtle" disabled={busy} onclick={()=>{selected=team.id;reason='';}}>{text($locale,'Түзетуге қайтару','Вернуть на исправление')}</button></div>
 {#if selected===team.id}<form onsubmit={event=>{event.preventDefault();void run(async()=>{await gateway.reject(team.id,reason);selected='';reason='';await load();notice=text($locale,'Өтінім себеппен қайтарылды.','Заявка возвращена с причиной.');});}}><label>{text($locale,'Не түзету керек?','Что нужно исправить?')}<textarea bind:value={reason} required minlength="8" maxlength="2000" rows="3"></textarea></label><div class="actions"><button class="button primary" disabled={busy}>{text($locale,'Себеппен қайтару','Вернуть с причиной')}</button><button type="button" class="button subtle" onclick={()=>selected=''}>{text($locale,'Бас тарту','Отмена')}</button></div></form>{/if}
 </article>{:else}{#if !error}<div class="empty-state"><h2>{text($locale,'Жаңа өтінім жоқ','Новых заявок нет')}</h2><p>{text($locale,'Жобаларға өтінімдер бөлек бөлімде қаралады.','Запросы на проекты рассматриваются в отдельном разделе.')}</p><a class="button subtle" href="/admin/requests">{text($locale,'Жоба өтінімдері','Заявки на проекты')}</a></div>{/if}{/each}
 {#if more}<button class="button subtle" disabled={busy} onclick={()=>run(()=>load(true))}>{text($locale,'Тағы өтінімдер','Ещё заявки')}</button>{/if}
 {/if}
{/if}
<style>article{max-width:840px;padding:24px 0;border-bottom:1px solid var(--line)}h2{font-size:20px;margin:0}p{white-space:pre-wrap;overflow-wrap:anywhere;line-height:1.6}small{color:var(--muted)}.actions{display:flex;gap:8px;flex-wrap:wrap;margin-top:16px}form{max-width:640px;margin-top:20px}label{display:grid;gap:8px}</style>
