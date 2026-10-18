<script lang="ts">
 import '../app.css';
 import {page} from '$app/stores';
 import {onMount,untrack} from 'svelte';
 import {createPreferences,createLocalePreferences,createNotifications,createAuth,createTeams} from '$lib/composition/browser';
 import {initializeTheme} from '$lib/features/preferences/application/theme';
 import {changeLocale} from '$lib/features/preferences/application/locale';
 import {provideLocale,text} from '$lib/shared/presentation/locale';
 import AppHeader from '$lib/shared/presentation/ui/AppHeader.svelte';
 import MobileNavigation from '$lib/shared/presentation/ui/MobileNavigation.svelte';
 import type {NavigationItem} from '$lib/shared/presentation/ui/types';
 import {goto,invalidateAll} from '$app/navigation';
 import AccountMenu from '$lib/features/preferences/presentation/AccountMenu.svelte';
 import SiteFooter from '$lib/features/preferences/presentation/SiteFooter.svelte';
 import {unreadBadge as badgeText} from '$lib/features/notifications/application/unread';
 import {avatarChanges} from '$lib/features/identity/infrastructure/avatar-changes';
 import {notificationChanges} from '$lib/features/notifications/infrastructure/notification-changes';
 import {safeReturnPath} from '$lib/features/identity/domain/return-path';
 let {children,data}=$props();
 const notifications=createNotifications(),auth=createAuth(),teams=createTeams();
 let avatarVersion=$state(0);
 let unreadCount=$state(0),mounted=$state(false),refreshGeneration=0;
 async function refreshUnread(){const generation=++refreshGeneration,id=data.user?.id; if(!id){unreadCount=0;return;}try{const count=await notifications.unread(id);if(generation===refreshGeneration&&data.user?.id===id)unreadCount=count;}catch{if(generation===refreshGeneration)unreadCount=0;}}
 onMount(()=>{mounted=true;const refresh=()=>{if(document.visibilityState==='visible')void refreshUnread();};const unsubscribe=notificationChanges.subscribe(refresh);const unsubscribeAvatar=avatarChanges.subscribe(actorId=>{if(actorId===data.user?.id)avatarVersion++;});window.addEventListener('focus',refresh);document.addEventListener('visibilitychange',refresh);return ()=>{mounted=false;refreshGeneration++;unsubscribe();unsubscribeAvatar();window.removeEventListener('focus',refresh);document.removeEventListener('visibilitychange',refresh);};});
 $effect(()=>{const id=data.user?.id;if(mounted){unreadCount=0;void refreshUnread();}});
 async function loggedOut(){await invalidateAll();await goto('/');}
 const unreadBadge=$derived(badgeText(data.user?.id||null,unreadCount));
 const adminArea=$derived($page.url.pathname==='/admin'||$page.url.pathname.startsWith('/admin/'));
 const localeStore=createLocalePreferences();
 const {locale}=provideLocale(untrack(()=>data.locale),value=>changeLocale(localeStore,value));
 onMount(()=>{initializeTheme(createPreferences()); if($page.url.pathname==='/offline-reader'||$page.url.pathname==='/saved') locale.set(localeStore.read()); localeStore.apply($locale);});
 $effect(()=>{localeStore.apply($locale);});
 const navigation=$derived<NavigationItem[]>([
  {label:text($locale,'Басты бет','Главная'),href:'/',icon:'home'},
  {label:'Каталог',href:'/catalog',icon:'catalog'},
  {label:text($locale,'Кітапхана','Библиотека'),shortLabel:text($locale,'Сөре','Списки'),href:'/library',icon:'bookmark'},
  {label:text($locale,'Хабарлар','Уведомления'),shortLabel:text($locale,'Хабарлар','События'),href:'/notifications',icon:'bell'},
  {label:text($locale,'Мәзір','Меню'),href:'/menu',icon:'menu'}
 ]);
 const loginReturn=$derived($page.url.pathname==='/saved'?'/saved':safeReturnPath(/^\/(login|register|forgot-password|reset-password|verify-email)(\/|$)/.test($page.url.pathname)?$page.url.searchParams.get('returnTo'):$page.url.pathname+$page.url.search));
</script>
<a class="skip-link" href="#content">{text($locale,'Мазмұнға өту','К содержимому')}</a>
{#if $page.url.pathname.startsWith('/read/') || $page.url.pathname==='/offline-reader'}
 <main id="content">{@render children()}</main>
{:else}
 {#if adminArea}<main id="content" class="admin-container">{@render children()}</main>{:else}
 <AppHeader {unreadBadge} {unreadCount} items={navigation.slice(0,4)} currentPath={$page.url.pathname} accountHref={data.user?'/menu':'/login?returnTo='+encodeURIComponent(loginReturn)} accountLabel={data.user?'Аккаунт':text($locale,'Кіру','Войти')} homeLabel={text($locale,'Qmanga басты бет','Qmanga — главная')} navigationLabel={text($locale,'Негізгі','Основная навигация')}>{#snippet account()}{#if data.user}<AccountMenu version={avatarVersion} user={data.user} {auth} {teams} onlogout={loggedOut}/>{:else}<a class="guest-login" href={'/login?returnTo='+encodeURIComponent(loginReturn)}>{text($locale,'Кіру','Войти')}</a>{/if}{/snippet}</AppHeader>
 <main id="content" class="page-container">{@render children()}</main>
 <SiteFooter version={data.buildVersion}/>
 <MobileNavigation {unreadBadge} {unreadCount} items={navigation} currentPath={$page.url.pathname} label={text($locale,'Мобильді навигация','Мобильная навигация')}/>
 {/if}
{/if}

<style>.admin-container{min-height:100dvh}.guest-login{display:flex;align-items:center;justify-content:center;min-width:44px;min-height:44px;color:var(--text);font-size:14px}</style>
