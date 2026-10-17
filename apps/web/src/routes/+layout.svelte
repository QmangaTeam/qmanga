<script lang="ts">
 import '../app.css';
 import {page} from '$app/stores';
 import {onMount,untrack} from 'svelte';
 import {createPreferences,createLocalePreferences} from '$lib/composition/browser';
 import {initializeTheme} from '$lib/features/preferences/application/theme';
 import {changeLocale} from '$lib/features/preferences/application/locale';
 import {provideLocale,text} from '$lib/shared/presentation/locale';
 import AppHeader from '$lib/shared/presentation/ui/AppHeader.svelte';
 import MobileNavigation from '$lib/shared/presentation/ui/MobileNavigation.svelte';
 import type {NavigationItem} from '$lib/shared/presentation/ui/types';
 let {children,data}=$props();
 const localeStore=createLocalePreferences();
 const {locale}=provideLocale(untrack(()=>data.locale),value=>changeLocale(localeStore,value));
 onMount(()=>{initializeTheme(createPreferences()); if($page.url.pathname==='/offline-reader') locale.set(localeStore.read()); localeStore.apply($locale);});
 $effect(()=>{localeStore.apply($locale);});
 const navigation=$derived<NavigationItem[]>([
  {label:text($locale,'Басты бет','Главная'),href:'/',icon:'home'},
  {label:'Каталог',href:'/catalog',icon:'catalog'},
  {label:text($locale,'Кітапхана','Библиотека'),href:'/library',icon:'bookmark'},
  {label:text($locale,'Хабарлар','Уведомления'),href:'/notifications',icon:'bell'},
  {label:text($locale,'Мәзір','Меню'),href:'/settings',icon:'menu'}
 ]);
</script>
<a class="skip-link" href="#content">{text($locale,'Мазмұнға өту','К содержимому')}</a>
{#if $page.url.pathname.startsWith('/read/') || $page.url.pathname==='/offline-reader'}
 <main id="content">{@render children()}</main>
{:else}
 <AppHeader items={navigation.slice(0,4)} currentPath={$page.url.pathname} accountHref={data.user?'/settings':'/login'} accountLabel={data.user?'Аккаунт':text($locale,'Кіру','Войти')} homeLabel={text($locale,'Qmanga басты бет','Qmanga — главная')} navigationLabel={text($locale,'Негізгі','Основная навигация')}/>
 <main id="content" class="page-container">{@render children()}</main>
 <footer class="site-footer"><span>Qmanga</span><a href="/settings">{text($locale,'Қазақша · Баптаулар','Русский · Настройки')}</a></footer>
 <MobileNavigation items={navigation} currentPath={$page.url.pathname} label={text($locale,'Мобильді навигация','Мобильная навигация')}/>
{/if}
