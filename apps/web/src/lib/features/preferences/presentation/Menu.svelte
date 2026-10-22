<script lang="ts">
 import {onMount,onDestroy} from 'svelte';
 import type {Account} from '$lib/features/identity/domain/account';
 import type {AuthGateway} from '$lib/features/identity/application/auth';
 import type {TeamsGateway} from '$lib/features/teams/application/teams';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import MenuTile from '$lib/shared/presentation/ui/MenuTile.svelte';
 let {user,teams,auth,onlogout,avatarVersion=0}:{user:Account|null;avatarVersion?:number;teams?:TeamsGateway;auth?:AuthGateway;onlogout?:()=>Promise<void>}=$props();
 const {locale}=useLocale();
 let translator=$state(false),teamLoading=$state(false),teamFailed=$state(false),busy=$state(false),logoutFailed=$state(false),avatarFailed=$state(false);
 let generation=0;
 $effect(()=>{user?.avatarURL;avatarVersion;avatarFailed=false;});
 onDestroy(()=>{generation++;});
 async function loadTeams(){
  if(!user?.verified||!teams)return;
  const sequence=++generation,actor=user.id;teamLoading=true;teamFailed=false;translator=false;
  try{for(let offset=0;;offset+=100){const page=await teams.list(offset);if(sequence!==generation||user?.id!==actor)return;if(page.some(team=>team.approved&&!!team.role)){translator=true;break;}if(page.length<100)break;}}
  catch{if(sequence===generation)teamFailed=true;}
  finally{if(sequence===generation)teamLoading=false;}
 }
 onMount(()=>{void loadTeams();});
 async function logout(){if(busy||!auth)return;busy=true;logoutFailed=false;try{const result=await auth.logout();if(result.ok)await onlogout?.();else logoutFailed=true;}catch{logoutFailed=true;}finally{busy=false;}}
</script>
<svelte:head><title>Аккаунт — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="menu-page">
 <h1>Аккаунт</h1>
 {#if user}
  <a class="identity" href={'/users/'+encodeURIComponent(user.id)}>
   <span class="avatar" aria-hidden="true">{#if user.avatarURL&&!avatarFailed}<img src={user.avatarURL+'?v='+avatarVersion} alt="" onerror={()=>avatarFailed=true}/>{:else}{(user.displayName||'Q').slice(0,2).toUpperCase()}{/if}</span>
   <span><strong>{user.displayName||'Аккаунт'}</strong><small>{text($locale,'Менің профилім','Мой профиль')}</small></span>
  </a>
 {:else}
  <p class="hint">{text($locale,'Кітапханаңыз бен аккаунтыңызға кіру.','Войдите, чтобы открыть свою библиотеку и аккаунт.')}</p>
  <MenuTile href="/login?returnTo=%2Fmenu" label={text($locale,'Кіру','Войти')} icon="user" row/>
 {/if}
 <div class="menu-sections">
  <section aria-labelledby="account-options"><h2 id="account-options">{text($locale,'Жеке бөлім','Личный раздел')}</h2><div class="menu-rows">
   <MenuTile href="/saved" label={text($locale,'Жүктелгендер','Скачанное')} icon="download" row/>
   <MenuTile href="/settings" label={text($locale,'Баптаулар','Настройки')} icon="settings" row/>
   {#if user}<MenuTile href="/reports" label={text($locale,'Өтініштерім','Мои обращения')} icon="flag" row/>{/if}
  </div></section>
  {#if user?.verified}
   {#if teamLoading}<p class="hint" role="status">{text($locale,'Жұмыс кеңістіктері жүктелуде…','Загружаем рабочие пространства…')}</p>{/if}
   {#if teamFailed}<div role="alert"><p>{text($locale,'Командаларды жүктеу мүмкін болмады.','Не удалось загрузить команды.')}</p><button class="text-action" onclick={loadTeams}>{text($locale,'Қайталау','Повторить')}</button></div>{/if}
   {#if translator||user.role==='admin'||user.role==='moderator'}
    <section aria-labelledby="account-workspaces"><h2 id="account-workspaces">{text($locale,'Жұмыс кеңістіктері','Рабочие пространства')}</h2><div class="menu-rows">
     {#if translator}<MenuTile href="/studio/projects" label={text($locale,'Аудармашы кабинеті','Кабинет переводчика')} icon="upload" row/><MenuTile href="/studio/teams" label={text($locale,'Командаларым','Мои команды')} icon="users" row/>{/if}
     {#if user.role==='admin'}<MenuTile href="/admin" label={text($locale,'Басқару','Управление')} icon="shield" row/>{:else if user.role==='moderator'}<MenuTile href="/moderation" label="Модерация" icon="shield" row/>{/if}
    </div></section>
   {/if}
  {/if}
  <section aria-labelledby="account-help"><h2 id="account-help">{text($locale,'Көмек және сайт туралы','Помощь и о сайте')}</h2><div class="menu-rows"><MenuTile href="/help" label={text($locale,'Көмек','Помощь')} icon="help" row/><MenuTile href="/legal/terms" label={text($locale,'Пайдалану шарттары','Условия использования')} icon="file" row/><MenuTile href="/legal/privacy" label={text($locale,'Жеке деректер','Конфиденциальность')} icon="shield" row/><MenuTile href="/rights" label={text($locale,'Құқық иелеріне','Правообладателям')} icon="flag" row/></div></section>
 </div>
 {#if user&&auth}<button class="text-action logout" disabled={busy} onclick={logout}>{text($locale,'Шығу','Выйти')}</button>{#if logoutFailed}<p role="alert">{text($locale,'Шығу орындалмады. Қайталап көріңіз.','Не удалось выйти. Повторите попытку.')}</p>{/if}{/if}
</div>
<style>
 .menu-page{max-width:640px;margin-inline:auto}h1{font-size:24px;margin-bottom:20px}.identity{display:flex;align-items:center;gap:14px;min-height:76px;margin-bottom:24px;color:var(--text);text-decoration:none}.identity>span:last-child{min-width:0}.identity strong{display:block;font-size:18px;overflow-wrap:anywhere}.identity small{display:block;color:var(--accent);margin-top:5px;font-size:13px}.avatar{width:52px;height:52px;flex:none;display:grid;place-items:center;overflow:hidden;border-radius:50%;background:var(--raised);color:var(--accent);font-weight:650}.avatar img{width:100%;height:100%;object-fit:cover}.menu-sections{display:grid;gap:24px;margin-top:24px}h2{font-size:14px;font-weight:550;color:var(--muted);margin-bottom:10px}.menu-rows{display:grid;gap:8px}.hint{color:var(--muted);font-size:14px;margin-bottom:16px}.text-action{min-height:44px;padding:10px 16px;color:var(--text);background:var(--panel);border:1px solid var(--line);border-radius:8px;cursor:pointer}.logout{margin-top:24px;width:100%;text-align:left}.identity:focus-visible,.text-action:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
</style>
