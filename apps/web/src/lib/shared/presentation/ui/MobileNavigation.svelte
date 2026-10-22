<script lang="ts">
 import {onMount} from 'svelte';
 import Icon from './Icon.svelte';
 import {useLocale,text} from '../locale';
 import type {NavigationItem} from './types';
 let {items,currentPath,label,unreadCount=0,unreadBadge=''}:{items:NavigationItem[];currentPath:string;label:string;unreadCount?:number;unreadBadge?:string}=$props();
 const {locale}=useLocale();
 let drawer:HTMLDialogElement,trigger=$state<HTMLButtonElement>(),ready=$state(false);
 const mobileItems=$derived(['/menu','/catalog','/','/library','/notifications'].flatMap(href=>items.filter(item=>item.href===href)));
 const drawerItems=$derived([...items.filter(item=>item.href!=='/menu'),{href:'/saved',label:text($locale,'Жүктелгендер','Скачанное'),icon:'download' as const},{href:'/help',label:text($locale,'Көмек','Помощь'),icon:'help' as const}]);
 function close(){drawer?.close();trigger?.focus({preventScroll:true});}
 function backdrop(event:MouseEvent){if(event.target!==drawer)return;const b=drawer.getBoundingClientRect();if(event.clientX<b.left||event.clientX>b.right||event.clientY<b.top||event.clientY>b.bottom)close();}
 $effect(()=>{currentPath;drawer?.close();});
 onMount(()=>{ready=true;const media=matchMedia('(min-width:768px)');const resize=()=>{if(media.matches)drawer?.close();};media.addEventListener('change',resize);return()=>media.removeEventListener('change',resize);});
</script>
<nav class="mobile-navigation" aria-label={label}>
 {#each mobileItems as item}
  {#if item.href==='/menu'}<button bind:this={trigger} disabled={!ready} type="button" aria-haspopup="dialog" onclick={()=>drawer.showModal()} aria-label={item.label}><Icon name="menu"/><span>{item.label}</span></button>
  {:else}<a href={item.href} aria-label={item.href==='/notifications'&&unreadBadge?item.label+': '+unreadCount:item.label} aria-current={currentPath===item.href?'page':undefined}><span class="nav-icon">{#if item.href==='/'}<img src="/favicon.png" alt="" width="24" height="24"/>{:else}<Icon name={item.icon}/>{/if}{#if item.href==='/notifications'&&unreadBadge}<b aria-hidden="true">{unreadBadge}</b>{/if}</span><span>{item.shortLabel||item.label}</span></a>{/if}
 {/each}
</nav>
<dialog bind:this={drawer} class="navigation-drawer" aria-labelledby="navigation-drawer-title" onclick={backdrop}>
 <header><h2 id="navigation-drawer-title"><span>Q</span>manga.</h2><button type="button" class="close-drawer" aria-label={text($locale,'Жабу','Закрыть')} onclick={close}><Icon name="close"/></button></header>
 <nav aria-label={text($locale,'Сайт бөлімдері','Разделы сайта')}>{#each drawerItems as item}<a href={item.href} onclick={()=>drawer.close()} aria-current={currentPath===item.href?'page':undefined}><span class="drawer-icon"><Icon name={item.icon}/></span>{item.label}</a>{/each}</nav>
 <p>{text($locale,'Манга, манхва және маньхуа қазақ тілінде.','Манга, манхва и маньхуа на казахском языке.')}</p>
</dialog>
<style>
 .mobile-navigation{position:fixed;inset:auto 0 0;z-index:40;display:grid;grid-template-columns:repeat(5,minmax(0,1fr));padding:4px 8px calc(4px + env(safe-area-inset-bottom));background:var(--panel);border-top:1px solid var(--line)}
 .mobile-navigation a,.mobile-navigation button{display:flex;min-height:54px;min-width:0;flex-direction:column;align-items:center;justify-content:center;gap:4px;color:var(--muted);font:inherit;font-size:11px;line-height:1.2;text-decoration:none;background:none;border:0;border-radius:8px;cursor:pointer}.mobile-navigation [aria-current]{color:var(--accent);background:var(--raised)}.nav-icon{display:flex;position:relative}.nav-icon b{position:absolute;top:-6px;left:14px;background:var(--accent);color:var(--accent-ink);font-size:10px;min-width:16px;padding:1px 4px;border-radius:10px}
 .navigation-drawer{position:fixed;inset:0 auto 0 0;margin:0;width:min(320px,calc(100vw - 40px));height:100dvh;max-height:100dvh;max-width:none;padding:0;border:0;border-right:1px solid var(--line);border-radius:0;background:var(--bg);color:var(--text);overflow:auto;overscroll-behavior:contain;box-shadow:12px 0 48px #0005}.navigation-drawer::backdrop{background:#0008}.navigation-drawer header{display:flex;align-items:center;justify-content:space-between;min-height:84px;padding:18px 20px;background:var(--panel);border-bottom:1px solid var(--line)}h2{font-size:25px;letter-spacing:-1px}h2 span{color:var(--accent)}.close-drawer{width:44px;height:44px;display:grid;place-items:center;border:0;border-radius:10px;background:var(--accent);color:var(--accent-ink);cursor:pointer}.navigation-drawer nav{display:grid;gap:4px;padding:16px 12px}.navigation-drawer a{display:flex;align-items:center;gap:16px;min-height:60px;padding:8px 12px;border-radius:8px;color:var(--text);text-decoration:none;font-size:14px}.navigation-drawer a[aria-current],.navigation-drawer a:hover{background:var(--panel);color:var(--accent)}.drawer-icon{display:grid;place-items:center;width:38px;height:38px;border-radius:10px;background:var(--raised);color:var(--muted);flex:none}.navigation-drawer p{padding:20px 24px;border-top:1px solid var(--line);color:var(--muted);font-size:12px;line-height:1.6}.navigation-drawer a:focus-visible,.navigation-drawer button:focus-visible,.mobile-navigation a:focus-visible,.mobile-navigation button:focus-visible{outline:2px solid var(--accent);outline-offset:-2px}@media(min-width:768px){.mobile-navigation{display:none}}@media(prefers-reduced-motion:no-preference){.navigation-drawer[open]{animation:drawer-enter .18s ease-out}@keyframes drawer-enter{from{transform:translateX(-100%)}to{transform:translateX(0)}}}
</style>
