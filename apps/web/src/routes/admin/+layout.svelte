<script lang="ts">
 import {onMount} from 'svelte';
 import {page} from '$app/stores';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import Icon from '$lib/shared/presentation/ui/Icon.svelte';
 let {children}=$props();const {locale}=useLocale();let dialog:HTMLDialogElement;
 const items=$derived([
  {href:'/admin',label:text($locale,'Шолу','Обзор')},
  {href:'/admin/catalog',label:text($locale,'Тайтлдар және тараулар','Тайтлы и главы')},
  {href:'/admin/requests',label:text($locale,'Жоба өтінімдері','Заявки на проекты')},
  {href:'/admin/teams',label:text($locale,'Командалар','Команды')},
  {href:'/admin/taxonomy',label:text($locale,'Жанрлар, санаттар, жинақтар','Жанры, категории, подборки')},
  {href:'/admin/moderation',label:'Модерация'},
  {href:'/admin/accounts',label:text($locale,'Пайдаланушылар','Пользователи')},
  {href:'/admin/reports',label:text($locale,'Өтініштер','Обращения')},
  {href:'/admin/audit',label:text($locale,'Аудит журналы','Журнал аудита')}
 ]);
 function close(){dialog?.close();}
 onMount(()=>{const media=matchMedia('(min-width:1024px)');const resize=()=>{if(media.matches)close();};media.addEventListener('change',resize);return ()=>media.removeEventListener('change',resize);});
 function backdrop(event:MouseEvent){if(event.target===dialog){const b=dialog.getBoundingClientRect();if(event.clientX<b.left||event.clientX>b.right||event.clientY<b.top||event.clientY>b.bottom)close();}}
 $effect(()=>{const path=$page.url.pathname;close();});
</script>
{#snippet navigation()}
 <nav aria-label={text($locale,'Әкімші бөлімдері','Разделы управления')}>{#each items as item}<a href={item.href} aria-current={$page.url.pathname===item.href||item.href!=='/admin'&&$page.url.pathname.startsWith(item.href+'/')?'page':undefined} onclick={close}>{item.label}</a>{/each}</nav>
 <a class="back-site" href="/">← {text($locale,'Сайтқа оралу','Вернуться на сайт')}</a>
{/snippet}
<div class="admin-shell">
 <aside class="admin-sidebar"><a class="brand" href="/admin">Qmanga <span>{text($locale,'Басқару','Управление')}</span></a>{@render navigation()}</aside>
 <div class="admin-workspace"><header><button class="open-nav" onclick={()=>dialog.showModal()} aria-haspopup="dialog"><Icon name="menu"/>{text($locale,'Бөлімдер','Разделы')}</button><a href="/settings">{text($locale,'Аккаунт','Аккаунт')}</a></header><div class="admin-content">{@render children()}</div></div>
</div>
<dialog bind:this={dialog} aria-labelledby="admin-nav-title" onclick={backdrop}><div class="dialog-heading"><h2 id="admin-nav-title">{text($locale,'Басқару','Управление')}</h2><button aria-label={text($locale,'Жабу','Закрыть')} onclick={close}><Icon name="close"/></button></div>{@render navigation()}</dialog>
<style>
 .admin-shell{min-height:100dvh;background:var(--bg)}.admin-sidebar{display:none}.admin-workspace{min-width:0}header{display:flex;align-items:center;justify-content:space-between;gap:16px;min-height:64px;padding:8px 16px;border-bottom:1px solid var(--line);background:var(--panel)}header a{color:var(--muted);font-size:14px}.open-nav,.dialog-heading button{display:flex;align-items:center;gap:8px;justify-content:center;background:none;border:0;color:var(--text);padding:0 8px;min-height:44px;min-width:44px}a{min-height:44px;display:flex;align-items:center;text-decoration:none}.admin-content{padding:24px 16px 48px;min-width:0;max-width:1500px}nav{display:grid;gap:4px}nav a{padding:8px 12px;color:var(--muted);font-size:14px;border-radius:8px}nav a[aria-current]{color:var(--text);background:var(--raised);box-shadow:inset 3px 0 var(--accent)}nav a:hover{color:var(--text)}.back-site{padding:12px;color:var(--muted);font-size:13px;margin-top:20px}.brand{font-size:22px;font-weight:700;display:grid;padding:4px 12px 24px;color:var(--text)}.brand span{font-size:12px;color:var(--muted);font-weight:500}
 dialog{position:fixed;inset:0 auto 0 0;margin:0;width:min(300px,calc(100vw - 24px));height:100dvh;max-height:100dvh;padding:12px;border:0;border-right:1px solid var(--line);background:var(--panel);color:var(--text);overflow:auto}.dialog-heading{display:flex;align-items:center;justify-content:space-between;padding:0 4px 16px}h2{font-size:17px}dialog::backdrop{background:#0009}dialog[open]{animation:nav-enter 200ms ease-out}@keyframes nav-enter{from{transform:translateX(-100%)}to{transform:translateX(0)}}@media(prefers-reduced-motion:reduce){dialog[open]{animation:none}}
 @media(min-width:1024px){.admin-shell{display:grid;grid-template-columns:256px minmax(0,1fr)}.admin-sidebar{display:block;position:sticky;top:0;height:100dvh;overflow:auto;padding:20px 12px;border-right:1px solid var(--line);background:var(--panel)}.open-nav{display:none}header{justify-content:flex-end;padding-inline:28px}.admin-content{padding-inline:28px}}
</style>
