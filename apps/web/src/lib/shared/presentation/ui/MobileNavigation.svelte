<script lang="ts">
 import Icon from './Icon.svelte';
 import type { NavigationItem } from './types';
 let { items, currentPath, label }: {items:NavigationItem[];currentPath:string;label:string} = $props();
</script>
<nav aria-label={label}>{#each items as item}<a href={item.href} aria-current={currentPath===item.href?'page':undefined}><Icon name={item.icon}/><span>{item.label}</span></a>{/each}</nav>
<style>
 nav { position:fixed; inset:auto 12px calc(8px + env(safe-area-inset-bottom)); z-index:40; display:grid; grid-template-columns:repeat(5,minmax(0,1fr)); background:var(--panel); border:1px solid var(--line); border-radius:16px; padding:4px; box-shadow:0 4px 16px #0003; } a { display:flex; min-height:52px; min-width:0; flex-direction:column; align-items:center; justify-content:center; gap:4px; text-decoration:none; color:var(--muted); font-size:11px; line-height:1.2; border-radius:10px; } a[aria-current] { color:var(--accent); background:var(--raised); } span { max-width:100%; overflow-wrap:anywhere; text-align:center; } @media(min-width:768px) { nav { display:none; } }
</style>
