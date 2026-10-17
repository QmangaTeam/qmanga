<script lang="ts">
 import {onMount,untrack} from 'svelte';
 import {page} from '$app/stores';
 import type {Account} from '$lib/features/identity/domain/account';
 import type {PublishingGateway} from '../application/publishing';
 import type {TeamsGateway} from '$lib/features/teams/application/teams';
 import type {ProjectRequestsGateway} from '../application/project-requests';
 import type {Team} from '$lib/features/teams/domain/team';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureCode,noticeText} from '$lib/shared/presentation/copy';
 import {ProjectDirectory} from '../application/project-directory';

 let {user,gateway,teamsGateway,requestsGateway}:{user:Account|null;gateway:PublishingGateway;teamsGateway:TeamsGateway;requestsGateway:ProjectRequestsGateway}=$props();
 const {locale}=useLocale();
 let teams=$state<Team[]>([]),team=$state(''),error=$state(''),query=$state('');
 let directoryVersion=$state(0);
 const directory=new ProjectDirectory(untrack(()=>gateway),()=>directoryVersion++);
 const items=$derived.by(()=>{directoryVersion;return directory.items;});
 const cursor=$derived.by(()=>{directoryVersion;return directory.cursor;});
 const hasMore=$derived.by(()=>{directoryVersion;return directory.hasMore;});
 const total=$derived.by(()=>{directoryVersion;return directory.total;});
 const loading=$derived.by(()=>{directoryVersion;return directory.loading;});
 const appliedCriteria=$derived.by(()=>{directoryVersion;return directory.appliedCriteria;});
 const directoryError=$derived.by(()=>{directoryVersion;return directory.error;});
 let searchGeneration=0;

 async function loadTeams(generation:number){
  try{
   const next=await teamsGateway.list();
   if(generation===searchGeneration){teams=next;error='';}
  }catch(failure){
   if(generation===searchGeneration)error=failureCode(failure);
  }
 }

 async function search(){
  const generation=++searchGeneration;
  error='';
  await directory.search(query,team);
  if(generation===searchGeneration&&!teams.length)await loadTeams(generation);
 }

 async function retry(){
  if(directory.error)await directory.retry();
  if(error)await loadTeams(searchGeneration);
 }

 onMount(()=>{
  team=$page.url.searchParams.get('team')||'';
  if(user)void search();
  return ()=>{searchGeneration++;};
 });
</script>
<svelte:head><title>{text($locale,'Аударма жобалары','Проекты переводов')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Жобалар','Проекты')} description={text($locale,'Тайтлды аш, тарау қос және аудармаңды жариялауға дайында.','Открой тайтл, добавь главу и подготовь перевод к публикации.')}>{#snippet actions()}{#if user}<a class="button primary" href="/studio/requests?new=1">{text($locale,'Жоба сұрау','Запросить проект')}</a>{/if}{/snippet}</SectionHeading>
{#if !user}<div class="empty-state"><h2>{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</h2><p>{text($locale,'Аккаунтқа кіріп, командаңа қосыл немесе жаңа командаға өтінім бер.','Войди и присоединись к команде или подай заявку на создание новой.')}</p><a class="button primary" href="/login?returnTo=%2Fstudio%2Fprojects">{text($locale,'Кіру','Войти')}</a></div>
{:else}
 {#if !user.verified}<div class="next-step"><h2>{text($locale,'Поштаңды раста','Подтверди почту')}</h2><p>{text($locale,'Командаға өтінім беру және файл жүктеу үшін пошта расталуы керек.','Для заявки команды и загрузки файлов нужна подтверждённая почта.')}</p><a class="button subtle" href="/settings">{text($locale,'Аккаунт баптаулары','Настройки аккаунта')}</a></div>{/if}
 {#if error || directoryError}<p role="alert" class="form-message">{noticeText($locale,directoryError || error)}</p><button class="button subtle" disabled={loading} onclick={()=>retry()}>{text($locale,'Қайталау','Повторить')}</button>{/if}
 <a class="pending" href="/studio/requests">{text($locale,'Өтінімдерім','Мои заявки')} <span>→</span></a>
 <form class="project-search" onsubmit={event=>{event.preventDefault();void search();}}><label>{text($locale,'Жобадан іздеу','Поиск проектов')}<input type="search" bind:value={query} maxlength="80"/></label>{#if teams.length>1}<label>{text($locale,'Команда','Команда')}<select aria-label={text($locale,'Команда','Команда')} bind:value={team} onchange={()=>search()}><option value="">{text($locale,'Барлық командалар','Все команды')}</option>{#each teams as item}<option value={item.id}>{item.name}</option>{/each}</select></label>{/if}<button class="button subtle" type="submit">{text($locale,'Іздеу','Найти')}</button></form>
 {#if total!==undefined}<p class="muted">{text($locale,'Табылды: ','Найдено: ')}{total}</p>{/if}
 {#if loading&&!items.length}<p role="status">{text($locale,'Жобалар жүктелуде…','Загружаем проекты…')}</p>
 {:else if !items.length&&!error&&!directoryError&&(appliedCriteria.teamId||appliedCriteria.query)}<p class="empty-state">{text($locale,'Жоба табылмады. Сұрау не команданы өзгерт.','Проекты не найдены. Измени запрос или команду.')}</p>
 {:else if !items.length&&!error&&!directoryError}<div class="next-step"><h2>{teams.some(team=>team.approved)?text($locale,'Бірінші жобаңды сұра','Запроси первый проект'):text($locale,'Командадан баста','Начни с команды')}</h2><p>{teams.some(team=>team.approved)?text($locale,'Каталогтан тайтл таңда немесе жаңа шығарма ұсын. Мақұлданған жоба осында пайда болады.','Выбери тайтл из каталога или предложи новый. Одобренный проект появится здесь.'):text($locale,'Команда құруға өтінім бер. Мақұлдағаннан кейін тайтл сұрап, тараулар жүктей аласың.','Подай заявку на команду. После одобрения можно запросить тайтл и загружать главы.')}</p><div class="actions"><a class="button primary" href={teams.some(team=>team.approved)?'/studio/requests?new=1':'/studio/teams?new=1'}>{teams.some(team=>team.approved)?text($locale,'Өтінім жіберу','Подать заявку'):text($locale,'Командаларға өту','Перейти к командам')}</a><a class="button subtle" href="/studio/requests">{text($locale,'Өтінімдерім','Мои заявки')}</a></div></div>
 {:else}<div class="projects">{#each items as item}<article><div><h2><a href={'/studio/projects/'+item.id}>{item.titleName}</a></h2><p>{item.teamName}</p></div><a class="button subtle" href={'/studio/projects/'+item.id}>{text($locale,'Тарауларды басқару','Управлять главами')}</a></article>{/each}</div>{/if}
 {#if hasMore}<button class="button subtle" disabled={loading} onclick={()=>directory.more()}>{text($locale,'Тағы жобалар','Ещё проекты')}</button>{/if}
{/if}
<style>
 .project-search{display:flex;align-items:end;flex-wrap:wrap;gap:12px;margin:20px 0}.project-search label{display:grid;gap:8px;flex:1 1 220px;min-width:0}
 .next-step{max-width:720px;padding:24px 0;border-block:1px solid var(--line);margin:20px 0}.next-step h2{font-size:20px;margin:0 0 12px}.next-step p{line-height:1.6;color:var(--muted);max-width:600px}.actions{display:flex;flex-wrap:wrap;gap:8px;margin-top:20px}.pending{display:flex;justify-content:space-between;align-items:center;min-height:48px;border-bottom:1px solid var(--line);padding:12px 0;margin-bottom:20px}.projects{max-width:900px}.projects article{display:flex;justify-content:space-between;align-items:center;gap:16px;padding:22px 0;border-bottom:1px solid var(--line)}article h2{font-size:18px;margin:0}article h2 a{display:inline-flex;align-items:center;min-height:44px;overflow-wrap:anywhere}article p{font-size:14px;color:var(--muted);margin:4px 0 0}article>div{min-width:0}@media(max-width:600px){.projects article{align-items:start;flex-direction:column;gap:12px}}
</style>
