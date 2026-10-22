<script lang="ts">
 import {page} from '$app/state';
 import type {Snippet} from 'svelte';
 import type {NavigationItem} from './types';
 import Icon from './Icon.svelte';
 import {useLocale,text} from '../locale';
 let {items,currentPath,accountHref,accountLabel,homeLabel,navigationLabel,account,theme,unreadCount=0,unreadBadge=''}:{items:NavigationItem[];currentPath:string;accountHref:string;accountLabel:string;homeLabel:string;navigationLabel:string;account?:Snippet;theme?:Snippet;unreadCount?:number;unreadBadge?:string}=$props();
 const {locale}=useLocale();
 const compact=$derived(/^\/(login|register|forgot-password|reset-password|verify-email|settings|notifications|users)(\/|$)/.test(currentPath));
 let searchDialog:HTMLDialogElement;
 let contribution:HTMLDetailsElement;
 $effect(()=>{currentPath;if(contribution)contribution.open=false;});
 let searchInput:HTMLInputElement;
 function openSearch(){searchInput.value=page.url.pathname==='/catalog'?(page.url.searchParams.get('q')||''):'';searchDialog.showModal();searchInput.focus();}
</script>
<header class:compact>
 <div class="header-inner">
  <a class="wordmark" href="/" aria-label={homeLabel}><img src="/favicon.png" alt="" width="40" height="40"/><span class="brand-name" aria-hidden="true">manga<span class="brand-dot">.</span></span></a>
  <nav class="primary-navigation" aria-label={navigationLabel}>
   <button class="search-trigger" onclick={openSearch}><Icon name="search"/><span>{text($locale,'Іздеу','Поиск')}</span></button>
   <a href="/catalog" aria-current={currentPath==='/catalog'?'page':undefined}><Icon name="catalog"/><span>Каталог</span></a>
   <a class="shelf-link" href="/library"><Icon name="bookmark"/><span>{text($locale,'Менің сөрем','Моя библиотека')}</span></a>
  </nav>
  <nav class="account-navigation" aria-label={text($locale,'Жеке бөлім','Личный раздел')}>
   {#each items.filter(item=>item.href==='/library'||item.href==='/notifications') as item}<a class="icon-link" href={item.href} aria-label={item.href==='/notifications'&&!!unreadBadge?item.label+': '+unreadCount:item.label} title={item.label} aria-current={currentPath===item.href?'page':undefined}><Icon name={item.icon}/>{#if item.href==='/notifications'&&!!unreadBadge}<span class="unread-badge" aria-hidden="true">{unreadBadge}</span>{/if}</a>{/each}
   <details bind:this={contribution} class="contribution-menu"><summary aria-label={text($locale,'Қосу','Добавить')} title={text($locale,'Аудармашыларға','Переводчикам')}><Icon name="plus"/></summary><nav aria-label={text($locale,'Аудармашыларға','Переводчикам')}><a href="/studio/projects">{text($locale,'Жобаларым','Мои проекты')}</a><a href="/studio/teams">{text($locale,'Командаларым','Мои команды')}</a><a href="/studio/requests">{text($locale,'Жобаға өтінім','Заявка на проект')}</a></nav></details>
   {#if theme}<span class="header-theme">{@render theme()}</span>{/if}
   {#if account}{@render account()}{:else}<a class="account-link" href={accountHref} aria-label={accountLabel} title={accountLabel}><Icon name="user"/></a>{/if}
  </nav>
 </div>
</header>
<dialog bind:this={searchDialog} aria-labelledby="global-search-title" onkeydown={event=>{if(event.key==='Escape'){event.preventDefault();searchDialog.close();}}}>
 <div class="search-heading"><h2 id="global-search-title">{text($locale,'Тайтл іздеу','Поиск тайтла')}</h2><button class="close-search" aria-label={text($locale,'Жабу','Закрыть')} onclick={()=>searchDialog.close()}><Icon name="close"/></button></div>
 <form action="/catalog" method="GET" role="search" onsubmit={()=>searchDialog.close()}>
  <label class="sr-only" for="global-query">{text($locale,'Тайтл іздеу','Поиск тайтла')}</label>
  <div class="search-field"><Icon name="search"/><input bind:this={searchInput} id="global-query" type="search" name="q" maxlength="80" required placeholder={text($locale,'Тайтл атауы','Название тайтла')}/><button type="submit">{text($locale,'Іздеу','Найти')}</button></div>
 </form>
 <a class="catalogue-link" href="/catalog" onclick={()=>searchDialog.close()}>{text($locale,'Каталог пен сүзгілерді ашу','Открыть каталог и фильтры')}<Icon name="right" size={16}/></a>
</dialog>
<style>
 header{position:sticky;top:0;z-index:30;background:var(--bg);border-bottom:1px solid var(--line)}
 .header-inner{max-width:1280px;min-height:60px;margin:auto;padding:8px 16px;display:flex;align-items:center;justify-content:space-between;gap:16px}
 .wordmark,.primary-navigation>a,.account-navigation>.icon-link{display:none}
 .account-navigation{display:flex;align-items:center;flex:none}.primary-navigation{min-width:0}
 .icon-link{position:relative}.unread-badge{position:absolute;top:0;right:0;min-width:18px;padding:1px 4px;border-radius:12px;background:var(--accent);color:var(--accent-ink);font-size:10px;line-height:16px}
 .primary-navigation{width:100%;display:flex;align-items:center;gap:24px}
 a,button{min-height:44px;min-width:44px;display:inline-flex;align-items:center;justify-content:center;gap:8px;color:var(--muted);font-size:14px;font-weight:550;text-decoration:none}
 button{border:0;background:none;padding:0 12px}
 .search-trigger{justify-content:flex-start;width:100%;background:var(--panel);border:1px solid var(--line);border-radius:10px}
 a:hover,button:hover,a[aria-current]{color:var(--text)}
 a[aria-current]{color:var(--accent)}
 dialog{position:fixed;inset:72px 16px auto;width:calc(100% - 32px);max-width:600px;margin:0 auto;padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel);color:var(--text);max-height:calc(100dvh - 96px);overflow:auto;box-shadow:0 16px 64px #0006}
 dialog::backdrop{background:#0008}
 .search-heading{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:16px}
 h2{font-size:18px;font-weight:650}
 .close-search{padding:0;border-radius:6px}
 .search-field{display:flex;align-items:center;gap:8px;border:1px solid var(--muted);border-radius:8px;padding-left:12px}
 .search-field:focus-within{outline:2px solid var(--accent);outline-offset:2px}
 .search-field input{flex:1;width:0;padding-inline:0;border:0;border-radius:0;background:none;outline:none}
 .search-field button{color:var(--accent);flex-shrink:0}
 .catalogue-link{margin-top:12px;font-weight:400}
 @media(max-width:767px){.compact .wordmark{display:inline-flex;align-items:center;gap:2px;flex:none;color:var(--text)}.compact .wordmark img{width:28px;height:28px}.compact .brand-name{font-size:22px;font-weight:850;letter-spacing:-1px;line-height:1}.compact .brand-dot{color:var(--accent)}.compact .primary-navigation{margin-left:auto;width:auto;flex:none}.compact .search-trigger{width:44px;padding:0;justify-content:center;background:none;border-color:transparent}.compact .search-trigger span{display:none}.compact .header-inner{gap:8px}}
 @media(min-width:768px){
  .header-inner{height:64px;padding:0 24px;display:grid;grid-template-columns:1fr auto 1fr}
  .wordmark{display:inline-flex;align-items:center;justify-self:start;gap:2px;color:var(--text)}.wordmark img{width:40px;height:40px;flex:none}
  .brand-name{font-size:30px;font-weight:850;letter-spacing:-1.4px;line-height:1;color:var(--text)}.brand-dot{color:var(--accent)}
  .primary-navigation{width:auto;gap:20px}.primary-navigation>a{display:inline-flex}
  .search-trigger{width:auto;background:var(--panel);border:1px solid var(--line);padding:0 12px}
  .account-navigation>.icon-link{display:inline-flex}.account-navigation{display:flex;justify-content:flex-end;align-items:center;gap:4px}
  .account-link{border:1px solid var(--line);border-radius:50%;background:var(--panel);margin-left:8px;color:var(--text)}
 }
 @media(min-width:1024px){}
 .contribution-menu{position:relative}.contribution-menu summary{display:grid;place-items:center;min-width:44px;min-height:44px;list-style:none;color:var(--muted);cursor:pointer;border-radius:8px}.contribution-menu summary::-webkit-details-marker{display:none}.contribution-menu nav{position:absolute;right:0;top:52px;width:240px;padding:8px;background:var(--panel);border:1px solid var(--line);border-radius:10px;box-shadow:0 12px 40px #0004;display:grid}.contribution-menu nav a{justify-content:flex-start;padding:8px 12px}.header-theme{display:none}@media(min-width:768px){.header-theme{display:contents}.account-navigation{gap:4px}.account-navigation>.icon-link:first-child{display:none}.primary-navigation .shelf-link{display:inline-flex}.search-trigger{min-width:180px}.header-inner{grid-template-columns:auto minmax(0,1fr) auto;gap:32px}.primary-navigation{justify-content:flex-end;gap:20px}.primary-navigation .search-trigger{order:3;margin-left:auto}}@media(max-width:767px){.contribution-menu{display:none}.header-inner{gap:12px}}
@media(min-width:768px) and (max-width:1023px){.account-navigation>.icon-link:first-child{display:inline-flex}.primary-navigation .shelf-link{display:none}.header-inner{gap:16px}.primary-navigation{gap:12px}.search-trigger{min-width:120px}}
</style>
