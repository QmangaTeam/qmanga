<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { createPreferences } from '$lib/composition/browser';
  import { initializeTheme } from '$lib/features/preferences/application/theme';
  onMount(() => { initializeTheme(createPreferences()); });
  let { children, data } = $props();
  const navigation = [
    { name: 'Басты бет', path: '/', icon: 'M3 10 12 3l9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1Z' },
    { name: 'Каталог', path: '/catalog', icon: 'M3 4h7v7H3z M14 4h7v7h-7z M3 15h7v6H3z M14 15h7v6h-7z' },
    { name: 'Кітапхана', path: '/library', icon: 'M5 3h14v19l-7-5-7 5Z' },
    { name: 'Хабарлар', path: '/notifications', icon: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9 M10 21h4' },
    { name: 'Мәзір', path: '/settings', icon: 'M4 6h16 M4 12h16 M4 18h16' }
  ];
</script>
<a class="skip-link" href="#content">Мазмұнға өту</a>
{#if $page.url.pathname.startsWith('/read/') || $page.url.pathname === '/offline-reader'}
 <main id="content">{@render children()}</main>
{:else}
<header class="topbar"><div class="topbar-inner"><a class="brand" href="/" aria-label="Qmanga басты бет"><span class="brand-mark">Q</span><span>qmanga<span class="brand-dot">.</span></span></a><nav class="desktop-nav" aria-label="Негізгі">{#each navigation.slice(0, 3) as item}<a class:active={$page.url.pathname === item.path} href={item.path}>{item.name}</a>{/each}</nav><a href={data.user ? '/settings' : '/login'} class="button subtle">{data.user ? 'Есептікжазба' : 'Кіру'} <span aria-hidden="true">↗</span></a></div></header>
<main id="content" class="page-container">{@render children()}</main>
<footer class="site-footer"><a class="brand" href="/">qmanga.</a><p>Әр оқиғаның өз тілі бар.</p></footer>
<nav class="bottom-nav" aria-label="Мобильді навигация">{#each navigation as item}<a class:active={$page.url.pathname === item.path} href={item.path} aria-current={$page.url.pathname === item.path ? 'page' : undefined}><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={item.icon}/></svg><span>{item.name}</span></a>{/each}</nav>
{/if}
