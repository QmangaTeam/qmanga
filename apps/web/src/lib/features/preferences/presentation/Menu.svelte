<script lang="ts">
 import type {Account} from '$lib/features/identity/domain/account';
 import type {IconName} from '$lib/shared/presentation/ui/types';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import MenuTile from '$lib/shared/presentation/ui/MenuTile.svelte';
 import ChoiceChips from '$lib/shared/presentation/ui/ChoiceChips.svelte';
 let {user}:{user:Account|null}=$props();
 const {locale,change}=useLocale();
 const reading=$derived<{href:string;label:string;icon:IconName}[]>([
  {href:'/catalog',label:'Каталог',icon:'catalog'},
  {href:'/library',label:text($locale,'Кітапхана','Библиотека'),icon:'bookmark'},
  {href:'/saved',label:text($locale,'Жүктелгендер','Скачанное'),icon:'download'},
  {href:'/notifications',label:text($locale,'Хабарлар','Уведомления'),icon:'bell'}
 ]);
 const studio=$derived<{href:string;label:string;icon:IconName}[]>([
  {href:'/studio/projects',label:text($locale,'Жобаларым','Мои проекты'),icon:'folder'},
  {href:'/studio/requests',label:text($locale,'Жобаға өтінім','Заявки на проекты'),icon:'file'},
  {href:'/studio/teams',label:text($locale,'Командаларым','Мои команды'),icon:'users'}
 ]);
 const staff=$derived<{href:string;label:string;icon:IconName}[]>(user?.role==='admin'||user?.role==='moderator'?[
  {href:'/moderation',label:'Модерация',icon:'shield'},
  ...(user.role==='admin'?[
   {href:'/admin/requests',label:text($locale,'Жоба өтінімдері','Заявки на проекты'),icon:'file' as const},
   {href:'/admin',label:text($locale,'Команда өтінімдері','Заявки команд'),icon:'users' as const},
   {href:'/admin/catalog',label:text($locale,'Каталогты басқару','Управление каталогом'),icon:'catalog' as const}
  ]:[])
 ]:[]);
</script>
<svelte:head><title>{text($locale,'Мәзір','Меню')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="menu-page">
 <div class="menu-heading"><h1>{text($locale,'Мәзір','Меню')}</h1><div class="language"><ChoiceChips label={text($locale,'Интерфейс тілі','Язык интерфейса')} value={$locale} options={[{value:'kk',label:'Қазақша'},{value:'ru',label:'Русский'}]} onchoose={value=>change(value==='ru'?'ru':'kk')}/></div></div>
 <div class="account-card"><MenuTile href={user?'/settings':'/login?returnTo=%2Fmenu'} label={user?(user.displayName||'Аккаунт'):text($locale,'Кіру','Войти')} icon="user" row/></div>
 <div class="menu-sections">
  <section><h2>{text($locale,'Оқу','Чтение')}</h2><div class="tile-grid">{#each reading as item}<MenuTile {...item}/>{/each}</div></section>
  <section><h2>{text($locale,'Аудармашыларға','Переводчикам')}</h2><div class="menu-rows">{#each studio as item}<MenuTile {...item} row/>{/each}</div></section>
  {#if staff.length}<section><h2>{text($locale,'Басқару','Управление')}</h2><div class="menu-rows">{#each staff as item}<MenuTile {...item} row/>{/each}</div></section>{/if}
  <section><h2>{text($locale,'Қосымша','Ещё')}</h2><div class="menu-rows"><MenuTile href="/settings" label={text($locale,'Баптаулар','Настройки')} icon="settings" row/><MenuTile href="/reports" label={text($locale,'Өтініштерім','Мои обращения')} icon="flag" row/><MenuTile href="/help" label={text($locale,'Көмек','Помощь')} icon="help" row/></div></section>
 </div>
</div>
<style>
 .menu-page{max-width:960px}.menu-heading{display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px;margin-bottom:16px}h1{font-size:24px}.language{width:176px;max-width:100%}.account-card{margin-bottom:24px;max-width:480px}.menu-sections{display:grid;gap:24px}h2{font-size:14px;font-weight:550;color:var(--muted);margin-bottom:10px}.tile-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}.menu-rows{display:grid;gap:8px}@media(min-width:768px){.menu-sections{grid-template-columns:repeat(2,minmax(0,1fr));gap:28px}.account-card{max-width:calc(50% - 14px)}}
</style>
