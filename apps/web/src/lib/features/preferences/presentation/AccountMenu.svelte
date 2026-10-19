<script lang="ts">
 import type {Account} from '$lib/features/identity/domain/account';
 import type {AuthGateway} from '$lib/features/identity/application/auth';
 import type {TeamsGateway} from '$lib/features/teams/application/teams';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {user,auth,teams,onlogout,version=0}:{user:Account;auth:AuthGateway;teams:TeamsGateway;onlogout:()=>Promise<void>;version?:number}=$props();
 const {locale}=useLocale();
 let dialog:HTMLDialogElement,trigger:HTMLButtonElement;
 let avatarFailed=$state(false);
 $effect(()=>{const url=user.avatarURL,currentVersion=version;avatarFailed=false;});
 let translator=$state(false),busy=$state(false),failed=$state(false);
 const initials=$derived((user.displayName||'Q').trim().slice(0,2).toUpperCase());
 function close(){dialog.close();trigger.focus();}
 async function open(){failed=false;dialog.showModal();try{translator=(await teams.list()).some(team=>team.approved&&!!team.role);}catch{translator=false;}}
 function backdrop(event:MouseEvent){if(event.target===dialog){const b=dialog.getBoundingClientRect();if(event.clientX<b.left||event.clientX>b.right||event.clientY<b.top||event.clientY>b.bottom)close();}}
 async function logout(){if(busy)return;busy=true;failed=false;try{const result=await auth.logout();if(result.ok){close();await onlogout();}else failed=true;}catch{failed=true;}finally{busy=false;}}
</script>
<button bind:this={trigger} class="avatar" aria-label={text($locale,'Аккаунт мәзірі','Меню аккаунта')} aria-haspopup="dialog" onclick={open}>{#if user.avatarURL&&!avatarFailed}<img src={user.avatarURL+'?v='+version} alt="" onerror={()=>avatarFailed=true}/>{:else}{initials}{/if}</button>
<dialog bind:this={dialog} class="account-menu" aria-labelledby="account-menu-title" onclick={backdrop}>
 <div class="heading"><h2 id="account-menu-title">{user.displayName||'Аккаунт'}</h2><button aria-label={text($locale,'Жабу','Закрыть')} onclick={close}><Icon name="close"/></button></div>
 <nav aria-label={text($locale,'Аккаунт әрекеттері','Действия аккаунта')}>
  <a href="/settings" onclick={close}><Icon name="settings"/>{text($locale,'Профиль және баптаулар','Профиль и настройки')}</a>
  <a href="/library" onclick={close}><Icon name="bookmark"/>{text($locale,'Кітапхана','Библиотека')}</a>
  <a href="/saved" onclick={close}><Icon name="download"/>{text($locale,'Сақталған тараулар','Сохранённые главы')}</a>
  {#if translator}<a href="/studio/projects" onclick={close}><Icon name="upload"/>{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</a>{/if}
  {#if user.role==='admin'&&user.verified}<a href="/admin" onclick={close}><Icon name="shield"/>{text($locale,'Басқару','Управление')}</a>{:else if user.role==='moderator'&&user.verified}<a href="/moderation" onclick={close}><Icon name="shield"/>Модерация</a>{/if}
  <button onclick={logout} disabled={busy}>{text($locale,'Шығу','Выйти')}</button>
 </nav>
 {#if failed}<p role="alert">{text($locale,'Шығу орындалмады. Қайталап көріңіз.','Не удалось выйти. Повторите попытку.')}</p>{/if}
</dialog>
<style>
 .avatar img{width:100%;height:100%;object-fit:cover;border-radius:50%}
 .avatar{width:44px;height:44px;flex:none;border:1px solid var(--line);border-radius:50%;background:var(--raised);color:var(--text);font-size:14px;font-weight:650}
 .account-menu{position:fixed;inset:72px 12px auto auto;margin:0;width:min(320px,calc(100vw - 24px));max-height:calc(100dvh - 96px);padding:12px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);box-shadow:0 16px 64px #0006;overflow:auto}
 .account-menu::backdrop{background:#0008}.heading{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:0 4px 8px}h2{font-size:17px;min-width:0;overflow-wrap:anywhere}button,a{min-height:44px}button{cursor:pointer}.heading button{display:grid;place-items:center;width:44px;min-width:44px;flex:none;background:none;color:var(--text);border:0;border-radius:8px}nav{display:grid}nav a,nav button{display:flex;align-items:center;gap:10px;padding:8px 12px;text-align:left;text-decoration:none;background:none;color:var(--text);border:0;border-radius:8px;font-size:14px}nav a:hover,nav button:hover{background:var(--raised)}p{padding:12px;color:var(--text)}
 @media(min-width:768px){.account-menu{right:max(24px,calc((100vw - 1232px)/2));width:300px}.account-menu::backdrop{background:transparent}}
</style>
