<script lang="ts">
 import type {Account} from '$lib/features/identity/domain/account';
 import type {AuthGateway} from '$lib/features/identity/application/auth';
 import type {TeamsGateway} from '$lib/features/teams/application/teams';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import type {ThemeStore} from '../application/theme';
 import {initializeTheme,changeTheme} from '../application/theme';
 import type {Theme} from '../domain/theme';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {user,auth,teams,onlogout,version=0,preferences}:{preferences?:ThemeStore;user:Account;auth:AuthGateway;teams:TeamsGateway;onlogout:()=>Promise<void>;version?:number}=$props();
 const {locale}=useLocale();
 let dialog:HTMLDialogElement,trigger:HTMLButtonElement;
 let avatarFailed=$state(false),theme=$state<Theme>('system');
 $effect(()=>{const url=user.avatarURL,currentVersion=version;avatarFailed=false;});
 let translator=$state(false),busy=$state(false),failed=$state(false);
 const initials=$derived((user.displayName||'Q').trim().slice(0,2).toUpperCase());
 function close(){dialog?.close();trigger?.focus();}
 async function open(){failed=false;if(preferences)theme=initializeTheme(preferences);dialog.showModal();try{translator=(await teams.list()).some(team=>team.approved&&!!team.role);}catch{translator=false;}}
 function backdrop(event:MouseEvent){if(event.target===dialog){const b=dialog.getBoundingClientRect();if(event.clientX<b.left||event.clientX>b.right||event.clientY<b.top||event.clientY>b.bottom)close();}}
 async function logout(){if(busy)return;busy=true;failed=false;try{const result=await auth.logout();if(result.ok){close();await onlogout();}else failed=true;}catch{failed=true;}finally{busy=false;}}
</script>
<button bind:this={trigger} class="avatar" aria-label={text($locale,'Аккаунт мәзірі','Меню аккаунта')} aria-haspopup="dialog" onclick={open}>{#if user.avatarURL&&!avatarFailed}<img src={user.avatarURL+'?v='+version} alt="" onerror={()=>avatarFailed=true}/>{:else}{initials}{/if}</button>
<dialog bind:this={dialog} class="account-menu" aria-labelledby="account-menu-title" onclick={backdrop}>
 <div class="heading"><span class="menu-avatar">{#if user.avatarURL&&!avatarFailed}<img src={user.avatarURL+'?v='+version} alt="" onerror={()=>avatarFailed=true}/>{:else}{initials}{/if}</span><div class="menu-identity"><h2 id="account-menu-title">{user.displayName||'Аккаунт'}</h2><p>{user.role==='admin'?text($locale,'Әкімші','Администратор'):user.role==='moderator'?text($locale,'Модератор','Модератор'):text($locale,'Оқырман','Читатель')}</p></div><button aria-label={text($locale,'Жабу','Закрыть')} onclick={close}><Icon name="close"/></button></div>
 <nav aria-label={text($locale,'Аккаунт әрекеттері','Действия аккаунта')}>
  <a class="profile-link" href={'/users/'+encodeURIComponent(user.id)} onclick={close}><Icon name="user"/><span>{text($locale,'Менің профилім','Мой профиль')}</span><Icon name="right" size={16}/></a>
  <a href="/library" onclick={close}><Icon name="bookmark"/>{text($locale,'Кітапхана','Библиотека')}</a>
  <a href="/saved" onclick={close}><Icon name="download"/>{text($locale,'Сақталған тараулар','Сохранённые главы')}</a>
  <a href="/settings" onclick={close}><Icon name="settings"/>{text($locale,'Баптаулар','Настройки')}</a>
  {#if translator}<a href="/studio/projects" onclick={close}><Icon name="upload"/>{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</a>{/if}
  {#if user.role==='admin'&&user.verified}<a href="/admin" onclick={close}><Icon name="shield"/>{text($locale,'Басқару','Управление')}</a>{:else if user.role==='moderator'&&user.verified}<a href="/moderation" onclick={close}><Icon name="shield"/>Модерация</a>{/if}

 </nav>
 {#if preferences}<div class="menu-appearance"><span>{text($locale,'Көрініс','Внешний вид')}</span><div role="group" aria-label={text($locale,'Тақырып','Тема')}>{#each [{value:'light',icon:'sun',kk:'Жарық',ru:'Светлая'},{value:'dark',icon:'moon',kk:'Қараңғы',ru:'Тёмная'},{value:'system',icon:'monitor',kk:'Жүйе',ru:'Системная'}] as choice}<button type="button" aria-label={text($locale,choice.kk,choice.ru)} aria-pressed={theme===choice.value} onclick={()=>theme=changeTheme(preferences!,choice.value)}><Icon name={choice.icon as 'sun'|'moon'|'monitor'} size={18}/></button>{/each}</div></div>{/if}
 <button class="logout" onclick={logout} disabled={busy}><Icon name="left"/>{text($locale,'Шығу','Выйти')}</button>
 {#if failed}<p role="alert">{text($locale,'Шығу орындалмады. Қайталап көріңіз.','Не удалось выйти. Повторите попытку.')}</p>{/if}
</dialog>
<style>
 .menu-avatar{width:40px;height:40px;display:grid;place-items:center;background:var(--raised);border:1px solid var(--line);border-radius:50%;flex:none;color:var(--accent);font-size:13px;font-weight:650;overflow:hidden}.menu-avatar img{width:100%;height:100%;object-fit:cover}.menu-identity{flex:1;min-width:0}.menu-identity h2{margin:0;font-size:14px}.menu-identity p{font-size:12px;color:var(--muted);margin:4px 0 0;padding:0}.profile-link span{flex:1}.menu-appearance{display:flex;align-items:center;justify-content:space-between;gap:12px;border-top:1px solid var(--line);margin:12px 0 4px;padding:12px 10px;font-size:13px}.menu-appearance>div{display:flex;gap:4px}.menu-appearance button{display:grid;place-items:center;width:44px;height:44px;background:var(--raised);border:1px solid var(--line);border-radius:8px;color:var(--muted)}.menu-appearance button[aria-pressed=true]{border-color:var(--accent);color:var(--accent)}.logout{display:flex;align-items:center;gap:10px;width:100%;border:0;background:none;color:var(--text);padding:8px 12px;font-size:14px;text-align:left}.account-menu a:focus-visible,.account-menu button:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}.avatar img{width:100%;height:100%;object-fit:cover;border-radius:50%}
 .avatar{width:44px;height:44px;flex:none;border:1px solid var(--line);border-radius:50%;background:var(--raised);color:var(--text);font-size:14px;font-weight:650}
 .account-menu{position:fixed;inset:72px 12px auto auto;margin:0;width:min(320px,calc(100vw - 24px));max-height:calc(100dvh - 96px);padding:12px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);box-shadow:0 16px 64px #0006;overflow:auto}
 .account-menu::backdrop{background:#0008}.heading{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:0 4px 8px}h2{font-size:17px;min-width:0;overflow-wrap:anywhere}button,a{min-height:44px}button{cursor:pointer}.heading button{display:grid;place-items:center;width:44px;min-width:44px;flex:none;background:none;color:var(--text);border:0;border-radius:8px}nav{display:grid}nav a{display:flex;align-items:center;gap:10px;padding:8px 12px;text-align:left;text-decoration:none;background:none;color:var(--text);border:0;border-radius:8px;font-size:14px}nav a:hover{background:var(--raised)}p{padding:12px;color:var(--text)}
 @media(min-width:768px){.account-menu{right:max(24px,calc((100vw - 1232px)/2));width:300px}.account-menu::backdrop{background:transparent}}
</style>
