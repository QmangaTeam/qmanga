<script lang="ts">
 import Icon from './Icon.svelte';
 import type { NavigationItem } from './types';
 let { items, currentPath, label }: {items:NavigationItem[];currentPath:string;label:string} = $props();
 const activeIndex = $derived(items.findIndex(item => currentPath === item.href));
</script>
<nav aria-label={label} style:--item-count={items.length}>
 <div class="active-indicator" aria-hidden="true" style:--active-index={Math.max(0, activeIndex)} class:visible={activeIndex >= 0}></div>
 {#each items as item}
  <a href={item.href} aria-label={item.label} aria-current={currentPath===item.href?'page':undefined}>
   <Icon name={item.icon}/><span>{item.shortLabel||item.label}</span>
  </a>
 {/each}
</nav>
<style>
 nav { position:fixed; inset:auto 12px calc(8px + env(safe-area-inset-bottom)); z-index:40; display:grid; grid-template-columns:repeat(var(--item-count),minmax(0,1fr)); background:var(--panel); border:1px solid var(--line); border-radius:16px; padding:4px; box-shadow:0 4px 16px #0003; }
 .active-indicator { position:absolute; top:4px; bottom:4px; left:4px; width:calc((100% - 8px) / var(--item-count)); border-radius:10px; background:var(--raised); pointer-events:none; opacity:0; transform:translateX(calc(var(--active-index) * 100%)); transition:transform 260ms cubic-bezier(.22,1,.36,1),opacity 160ms ease; }
 .active-indicator.visible { opacity:1; }
 a { position:relative; display:flex; min-height:52px; min-width:0; flex-direction:column; align-items:center; justify-content:center; gap:4px; text-decoration:none; color:var(--muted); font-size:11px; line-height:1.2; border-radius:10px; -webkit-tap-highlight-color:transparent; transition:color 180ms ease,transform 120ms ease; }
 a[aria-current] { color:var(--accent); }
 a:active { transform:scale(.94); }
 a:focus-visible { outline:2px solid var(--accent); outline-offset:-2px; }
 span { max-width:100%; white-space:nowrap; text-align:center; }
 @media(prefers-reduced-motion:reduce) { .active-indicator,a { transition:none; } a:active { transform:none; } }
 @media(min-width:768px) { nav { display:none; } }
</style>
