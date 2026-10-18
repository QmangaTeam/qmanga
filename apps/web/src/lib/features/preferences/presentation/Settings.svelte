<script lang="ts">
 import AvatarEditor from '$lib/features/identity/presentation/AvatarEditor.svelte';
 import type {AvatarGateway} from '$lib/features/identity/application/avatars';
 import {onMount} from 'svelte';
 import {goto,invalidateAll} from '$app/navigation';
 import type {AuthGateway,AuthResult} from '$lib/features/identity/application/auth';
 import type {Account} from '$lib/features/identity/domain/account';
 import type {ThemeStore} from '../application/theme';
 import {initializeTheme,changeTheme} from '../application/theme';
 import type {Theme} from '../domain/theme';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {errorText} from '$lib/shared/presentation/messages';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import ChoiceChips from '$lib/shared/presentation/ui/ChoiceChips.svelte';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {user,auth,preferences,avatars}:{user:Account|null;auth:AuthGateway;preferences:ThemeStore;avatars?:AvatarGateway}=$props();
 const {locale,change}=useLocale();
 let theme=$state<Theme>('system'),busy=$state(false),result=$state<AuthResult|null>(null);
 let displayName=$state(''),profileResult=$state<AuthResult|null>(null);
 onMount(()=>{theme=initializeTheme(preferences);displayName=user?.displayName||'';});
 async function saveProfile(){if(!user||busy)return;busy=true;try{profileResult=await auth.updateProfile(user.id,displayName);if(profileResult.ok)await invalidateAll();}finally{busy=false;}}
 async function logout(){if(busy)return;busy=true;try{result=await auth.logout();if(result.ok){await invalidateAll();await goto('/');}}finally{busy=false;}}
 async function resend(){if(!user||busy)return;busy=true;try{result=await auth.resend(user.email);}finally{busy=false;}}
</script>
<svelte:head><title>{text($locale,'Баптаулар','Настройки')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Баптаулар','Настройки')}/>
<div class="settings">
 <section><h2>{text($locale,'Сыртқы көрініс','Внешний вид')}</h2>
 <div class="appearance-setting" id="ui-language"><h3>{text($locale,'Интерфейс тілі','Язык интерфейса')}</h3><ChoiceChips label={text($locale,'Интерфейс тілі','Язык интерфейса')} value={$locale} options={[{value:'kk',label:'Қазақша'},{value:'ru',label:'Русский'}]} onchoose={value=>change(value==='ru'?'ru':'kk')}/></div>
 <div class="appearance-setting" id="ui-theme"><h3>{text($locale,'Тақырып','Тема')}</h3><ChoiceChips label={text($locale,'Тақырып','Тема')} value={theme} options={[{value:'system',label:text($locale,'Авто','Авто'),icon:'monitor'},{value:'dark',label:text($locale,'Қараңғы','Тёмная'),icon:'moon'},{value:'light',label:text($locale,'Жарық','Светлая'),icon:'sun'}]} onchoose={value=>theme=changeTheme(preferences,value)}/></div>
 </section>
 <section><h2>Аккаунт</h2>{#if user}<p class="account-email">{user.email}</p><p class="muted">{user.verified?text($locale,'Пошта расталған','Почта подтверждена'):text($locale,'Пошта расталмаған','Почта не подтверждена')}</p>
 {#if avatars}{#key user.id}<AvatarEditor gateway={avatars} kind="user" actorId={user.id} ownerId={user.id} name={user.displayName||'Q'}/>{/key}{/if}
 <form class="profile-form" onsubmit={event=>{event.preventDefault();void saveProfile();}}><label for="profile-name">{text($locale,'Көрінетін атың','Публичное имя')}</label><input id="profile-name" bind:value={displayName} maxlength="40" autocomplete="nickname" aria-describedby="profile-hint"/><p id="profile-hint" class="muted">{text($locale,'Пікірлерде және команда құрамында көрінеді. Поштаңды жазба.','Видно в комментариях и составе команды. Не указывай почту.')}</p><Button type="submit" {busy}>{text($locale,'Профильді сақтау','Сохранить профиль')}</Button></form>
 {#if profileResult}<StateMessage kind={profileResult.ok?'success':'error'} title={profileResult.ok?text($locale,'Профиль сақталды','Профиль сохранён'):errorText($locale,profileResult.code)}/>{/if}
 <div class="settings-links"><ActionLink href="/reports">{text($locale,'Менің өтініштерім','Мои обращения')}</ActionLink>{#if !user.verified}<Button {busy} onclick={resend}>{text($locale,'Растау хатын жіберу','Отправить подтверждение')}</Button>{/if}<ActionLink href="/studio">{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</ActionLink>{#if user.role==='admin'}<ActionLink href="/admin">{text($locale,'Әкімші панелі','Администрирование')}</ActionLink>{/if}{#if user.role==='admin'||user.role==='moderator'}<ActionLink href="/moderation">{text($locale,'Модерация','Модерация')}</ActionLink>{/if}<Button onclick={logout} {busy} variant="ghost">{text($locale,'Шығу','Выйти')}</Button></div>
 {:else}<p class="muted">{text($locale,'Кітапхана мен оқу орнын басқа құрылғыларда да сақтау үшін кір.','Войди, чтобы сохранить библиотеку и место чтения на других устройствах.')}</p><div class="settings-links"><ActionLink href="/login" variant="primary">{text($locale,'Кіру','Войти')}</ActionLink><ActionLink href="/register">{text($locale,'Тіркелу','Регистрация')}</ActionLink></div>{/if}
 {#if result}<StateMessage kind={result.ok?'success':'error'} title={result.ok?text($locale,'Дайын. Поштаңды тексер.','Готово. Проверь почту.'):errorText($locale,result.code)}/>{/if}</section>
 <section><h2>{text($locale,'Оқу','Чтение')}</h2><div class="settings-links"><ActionLink href="/saved">{text($locale,'Сақталған тараулар','Скачанные главы')}</ActionLink><ActionLink href="/library">{text($locale,'Кітапхана','Библиотека')}</ActionLink></div></section>
</div>
<style>.appearance-setting{display:grid;gap:10px;padding:16px 0;border-bottom:1px solid var(--line)}.appearance-setting h3{font-size:14px;font-weight:550}.profile-form{display:grid;gap:12px;margin-top:20px}.profile-form :global(button){justify-self:start}.settings { max-width:720px; } section { margin-bottom:32px; } h2 { margin-bottom:16px; } .settings-links { display:flex; flex-wrap:wrap; gap:8px; margin-top:16px; } .account-email { overflow-wrap:anywhere; margin-bottom:4px; }</style>
