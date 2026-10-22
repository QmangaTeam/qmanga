<script lang="ts">
 import Icon from './Icon.svelte';
 let avatarFailed=$state(false);
 $effect(()=>{avatarURL;avatarVersion;avatarFailed=false;});
 import type { NavigationItem } from './types';
 let { items, currentPath, label, unreadCount=0, unreadBadge='', avatarURL, avatarVersion=0 }: {items:NavigationItem[];currentPath:string;label:string;unreadCount?:number;unreadBadge?:string;avatarURL?:string;avatarVersion?:number} = $props();
 const mobileItems = $derived.by(() => {
  const home = items.find(item => item.href === '/');
  const destinations = items.filter(item => item.href !== '/');
  if (home) destinations.splice(Math.floor(destinations.length / 2), 0, home);
  return destinations;
 });
 const activeIndex = $derived(mobileItems.findIndex(item => currentPath === item.href));
</script>
<nav aria-label={label} style:--item-count={mobileItems.length}>
 <div class="active-indicator" aria-hidden="true" style:--active-index={Math.max(0, activeIndex)} class:visible={activeIndex >= 0}></div>
 {#each mobileItems as item}
  <a href={item.href} aria-label={item.href==='/notifications'&&!!unreadBadge?item.label+': '+unreadCount:item.label} aria-current={currentPath===item.href?'page':undefined}>
   <div class="nav-icon">{#if item.href==='/'}<img class="home-logo" src="/favicon.png" alt="" width="24" height="24"/>{:else if item.href==='/menu'&&avatarURL&&!avatarFailed}<img class="account-avatar" src={avatarURL+'?v='+avatarVersion} alt="" width="24" height="24" onerror={()=>avatarFailed=true}/>{:else}<Icon name={item.icon}/>{/if}{#if item.href==='/notifications'&&!!unreadBadge}<b class="unread-badge" aria-hidden="true">{unreadBadge}</b>{/if}</div><span>{item.shortLabel||item.label}</span>
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
 .account-avatar{width:24px;height:24px;border-radius:50%;object-fit:cover}.home-logo{display:block;flex:none}.nav-icon{position:relative;display:flex}.unread-badge{position:absolute;top:-6px;left:14px;min-width:18px;padding:1px 4px;border-radius:12px;background:var(--accent);color:var(--accent-ink);font-size:10px;line-height:16px}
 span { max-width:100%; white-space:nowrap; text-align:center; }
 @media(prefers-reduced-motion:reduce) { .active-indicator,a { transition:none; } a:active { transform:none; } }
 @media(min-width:768px) { nav { display:none; } }
</style>
