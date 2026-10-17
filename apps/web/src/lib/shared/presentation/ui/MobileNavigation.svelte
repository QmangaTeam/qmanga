<script lang="ts">
 import Icon from './Icon.svelte';
 import type { NavigationItem } from './types';
 let { items, currentPath, label }: {items:NavigationItem[];currentPath:string;label:string} = $props();
</script>
<nav aria-label={label}>{#each items as item}<a href={item.href} aria-current={currentPath===item.href?'page':undefined}><Icon name={item.icon}/><span>{item.label}</span></a>{/each}</nav>
<style>
 nav { position:fixed; inset:auto 0 0; z-index:40; display:grid; grid-template-columns:repeat(5,minmax(0,1fr)); background:var(--bg); border-top:1px solid var(--line); padding:4px 0 calc(4px + env(safe-area-inset-bottom)); } a { display:flex; min-height:52px; min-width:0; flex-direction:column; align-items:center; justify-content:center; gap:4px; text-decoration:none; color:var(--muted); font-size:12px; line-height:1.2; } a[aria-current] { color:var(--accent); } span { max-width:100%; overflow-wrap:anywhere; text-align:center; } @media(min-width:768px) { nav { display:none; } }
</style>
