<script lang="ts">
 import type { NavigationItem } from './types';
 import ActionLink from './ActionLink.svelte';
 let { items, currentPath, accountHref, accountLabel, homeLabel, navigationLabel }: {items:NavigationItem[]; currentPath:string; accountHref:string; accountLabel:string; homeLabel:string; navigationLabel:string} = $props();
</script>
<header><div class="header-inner"><a class="wordmark" href="/" aria-label={homeLabel}>qmanga<span>.</span></a><nav aria-label={navigationLabel}>{#each items as item}<a href={item.href} aria-current={currentPath===item.href?'page':undefined}>{item.label}</a>{/each}</nav><ActionLink href={accountHref} variant="ghost">{accountLabel}</ActionLink></div></header>
<style>
 header { position:sticky; top:0; z-index:30; border-bottom:1px solid var(--line); background:var(--bg); } .header-inner { max-width:1200px; height:56px; margin:auto; padding:0 16px; display:flex; align-items:center; gap:16px; justify-content:space-between; } .wordmark { display:inline-flex; align-items:center; min-height:44px; color:var(--text); font-size:23px; font-weight:750; letter-spacing:-1px; text-decoration:none; } .wordmark span { color:var(--accent); } nav { display:none; align-items:center; gap:24px; font-size:14px; } nav a { min-height:44px; display:flex; align-items:center; color:var(--muted); text-decoration:none; } nav a[aria-current] { color:var(--text); box-shadow:inset 0 -2px var(--accent); } @media(min-width:768px) { .header-inner { padding-inline:24px; } nav { display:flex; } }
</style>
