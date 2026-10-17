<script lang="ts">
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
 import SettingsRow from '$lib/shared/presentation/ui/SettingsRow.svelte';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {user,auth,preferences}:{user:Account|null;auth:AuthGateway;preferences:ThemeStore}=$props();
 const {locale,change}=useLocale();
 let theme=$state<Theme>('system'),busy=$state(false),result=$state<AuthResult|null>(null);
 onMount(()=>{theme=initializeTheme(preferences);});
 async function logout(){if(busy)return;busy=true;try{result=await auth.logout();if(result.ok){await invalidateAll();await goto('/');}}finally{busy=false;}}
 async function resend(){if(!user||busy)return;busy=true;try{result=await auth.resend(user.email);}finally{busy=false;}}
</script>
<svelte:head><title>{text($locale,'Баптаулар','Настройки')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<SectionHeading level={1} title={text($locale,'Баптаулар','Настройки')}/>
<div class="settings">
 <section><h2>{text($locale,'Сыртқы көрініс','Внешний вид')}</h2>
 <SettingsRow controlId="ui-language" label={text($locale,'Интерфейс тілі','Язык интерфейса')} description={text($locale,'Манга мен пікірлердің тілі өзгермейді.','Язык манги и комментариев не меняется.')}><select id="ui-language" value={$locale} onchange={event=>change(event.currentTarget.value==='ru'?'ru':'kk')}><option value="kk">Қазақша</option><option value="ru">Русский</option></select></SettingsRow>
 <SettingsRow controlId="ui-theme" label={text($locale,'Тақырып','Тема')}><select id="ui-theme" value={theme} onchange={event=>theme=changeTheme(preferences,event.currentTarget.value)}><option value="system">{text($locale,'Құрылғыдағыдай','Как на устройстве')}</option><option value="dark">{text($locale,'Қараңғы','Тёмная')}</option><option value="light">{text($locale,'Жарық','Светлая')}</option></select></SettingsRow>
 </section>
 <section><h2>Аккаунт</h2>{#if user}<p class="account-email">{user.email}</p><p class="muted">{user.verified?text($locale,'Пошта расталған','Почта подтверждена'):text($locale,'Пошта расталмаған','Почта не подтверждена')}</p>
 <div class="settings-links">{#if !user.verified}<Button {busy} onclick={resend}>{text($locale,'Растау хатын жіберу','Отправить подтверждение')}</Button>{/if}<ActionLink href="/studio">{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</ActionLink>{#if user.role==='admin'}<ActionLink href="/admin">{text($locale,'Әкімші панелі','Администрирование')}</ActionLink>{/if}{#if user.role==='admin'||user.role==='moderator'}<ActionLink href="/moderation">{text($locale,'Модерация','Модерация')}</ActionLink>{/if}<Button onclick={logout} {busy} variant="ghost">{text($locale,'Шығу','Выйти')}</Button></div>
 {:else}<p class="muted">{text($locale,'Кітапхана мен оқу орнын басқа құрылғыларда да сақтау үшін кір.','Войди, чтобы сохранить библиотеку и место чтения на других устройствах.')}</p><div class="settings-links"><ActionLink href="/login" variant="primary">{text($locale,'Кіру','Войти')}</ActionLink><ActionLink href="/register">{text($locale,'Тіркелу','Регистрация')}</ActionLink></div>{/if}
 {#if result}<StateMessage kind={result.ok?'success':'error'} title={result.ok?text($locale,'Дайын. Поштаңды тексер.','Готово. Проверь почту.'):errorText($locale,result.code)}/>{/if}</section>
 <section><h2>{text($locale,'Оқу','Чтение')}</h2><div class="settings-links"><ActionLink href="/saved">{text($locale,'Сақталған тараулар','Скачанные главы')}</ActionLink><ActionLink href="/library">{text($locale,'Кітапхана','Библиотека')}</ActionLink></div></section>
</div>
<style>.settings { max-width:720px; } section { margin-bottom:32px; } h2 { margin-bottom:16px; } .settings-links { display:flex; flex-wrap:wrap; gap:8px; margin-top:16px; } .account-email { overflow-wrap:anywhere; margin-bottom:4px; }</style>
