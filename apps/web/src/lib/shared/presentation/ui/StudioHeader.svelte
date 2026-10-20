<script lang="ts">
 import type {Snippet} from 'svelte';
 import type {NavigationItem} from './types';
 import {useLocale,text} from '../locale';
 import Icon from './Icon.svelte';
 let {items,account,unreadCount=0,unreadBadge=''}:{items:NavigationItem[];account:Snippet;unreadCount?:number;unreadBadge?:string}=$props();
 const {locale}=useLocale();
 let navigationDialog:HTMLDialogElement;
</script>
<header class="studio-header">
 <div class="studio-header-inner">
  <a class="site-return" href="/"><Icon name="left"/>{text($locale,'Сайтқа','На сайт')}</a>
  <span class="studio-label">{text($locale,'Аудармашы кабинеті','Кабинет переводчика')}</span>
  <div class="header-actions"><button class="site-navigation" aria-haspopup="dialog" aria-label={text($locale,'Сайт навигациясы','Навигация сайта')} title={text($locale,'Сайт навигациясы','Навигация сайта')} onclick={()=>navigationDialog.showModal()}><Icon name="menu"/>{#if unreadBadge}<b class="unread-badge" aria-hidden="true">{unreadBadge}</b>{/if}</button>{@render account()}</div>
 </div>
</header>
<dialog bind:this={navigationDialog} aria-labelledby="studio-site-navigation">
 <div class="dialog-heading"><h2 id="studio-site-navigation">{text($locale,'Сайт навигациясы','Навигация сайта')}</h2><button aria-label={text($locale,'Жабу','Закрыть')} onclick={()=>navigationDialog.close()}><Icon name="close"/></button></div>
 <nav aria-labelledby="studio-site-navigation">{#each items as item}<a href={item.href} onclick={()=>navigationDialog.close()}><Icon name={item.icon}/><span>{item.label}</span>{#if item.href==='/notifications'&&unreadBadge}<b class="unread-badge" aria-label={text($locale,'Оқылмаған: ','Непрочитанных: ')+unreadCount}>{unreadBadge}</b>{/if}</a>{/each}</nav>
</dialog>
<style>
 header{position:sticky;top:0;z-index:30;background:var(--bg);border-bottom:1px solid var(--line)}
 .studio-header-inner{max-width:1280px;min-height:60px;margin:auto;padding:8px 16px;display:flex;align-items:center;justify-content:space-between;gap:12px}
 .site-return{display:inline-flex;align-items:center;gap:8px;min-height:44px;color:var(--text);text-decoration:none;font-size:14px}
 .studio-label{display:none;color:var(--muted);font-size:14px}.header-actions{display:flex;align-items:center;gap:8px}
 button{display:inline-flex;align-items:center;justify-content:center;min-width:44px;min-height:44px;padding:8px;border:1px solid var(--line);border-radius:8px;background:var(--panel);color:var(--text)}
 .site-navigation{position:relative}.unread-badge{min-width:18px;padding:1px 4px;border-radius:12px;background:var(--accent);color:var(--accent-ink);font-size:10px;line-height:16px}.site-navigation .unread-badge{position:absolute;top:-4px;right:-4px}
 dialog{width:calc(100% - 32px);max-width:420px;max-height:calc(100dvh - 32px);overflow:auto;padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);box-shadow:0 16px 64px #0006}dialog::backdrop{background:#0008}
 .dialog-heading{display:flex;align-items:center;justify-content:space-between;gap:12px}.dialog-heading h2{font-size:18px;margin:0}
 nav{display:grid;gap:4px;margin-top:12px}nav a{display:flex;align-items:center;gap:12px;min-height:44px;padding:8px;color:var(--text);text-decoration:none;border-radius:6px}nav a:hover{background:var(--raised)}nav a span{flex:1}
 @media(min-width:640px){.studio-label{display:block}.studio-header-inner{padding-inline:24px}}
</style>
