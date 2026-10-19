<script lang="ts">
 import type {NavigationItem} from './types';
 let {items,currentPath,label,sidebar=false,menuLabel=''}:{items:Pick<NavigationItem,'href'|'label'>[];currentPath:string;label:string;sidebar?:boolean;menuLabel?:string}=$props();
 let expanded=$state(false);
 const current=$derived(items.find(item=>active(item.href))?.label);
 $effect(()=>{currentPath;expanded=false;});
 function active(href:string){return href==='/admin'||href==='/moderation'?currentPath===href:currentPath===href||currentPath.startsWith(href+'/');}
</script>
<div class:sidebar class="workspace-navigation">
 {#if sidebar}<button class="section-toggle" type="button" aria-expanded={expanded} aria-controls="workspace-sections" onclick={()=>expanded=!expanded}><span>{menuLabel}</span><strong>{current}</strong><span aria-hidden="true">{expanded?'−':'+'}</span></button>{/if}
 <nav id={sidebar?'workspace-sections':undefined} class="workspace-nav" class:expanded aria-label={label}>{#each items as item}<a href={item.href} aria-current={active(item.href)?'page':undefined}>{item.label}</a>{/each}</nav>
</div>
<style>
 .workspace-nav{display:flex;gap:4px;overflow-x:auto;border-bottom:1px solid var(--line);margin:0 0 24px;scrollbar-width:thin}
 a{display:flex;align-items:center;min-height:48px;padding:0 16px;white-space:nowrap;color:var(--muted);font-size:14px;text-decoration:none;border-bottom:2px solid transparent}
 a[aria-current]{color:var(--text);border-color:var(--accent);font-weight:650}a:focus-visible,.section-toggle:focus-visible{outline:2px solid var(--accent);outline-offset:-3px}
 .section-toggle{display:flex;align-items:center;gap:12px;width:100%;min-height:48px;padding:12px 14px;color:var(--text);background:var(--panel);border:1px solid var(--line);border-radius:10px;cursor:pointer;text-align:start;font:inherit;font-size:14px}
 .section-toggle>span:first-child{color:var(--muted)}.section-toggle strong{flex:1;font-weight:600}
 .sidebar .workspace-nav{display:none;margin:8px 0 0;border:1px solid var(--line);border-radius:10px;background:var(--panel);padding:6px;flex-direction:column;overflow:visible}
 .sidebar .workspace-nav.expanded{display:flex}.sidebar a{white-space:normal;border:0;border-radius:7px;min-height:44px;padding:10px 12px}.sidebar a[aria-current]{background:var(--bg);box-shadow:inset 3px 0 var(--accent)}
 @media(min-width:900px){.sidebar .section-toggle{display:none}.sidebar{position:sticky;top:88px}.sidebar .workspace-nav{display:flex;margin:0}}
 @media(max-width:420px){a{padding-inline:12px}}
</style>
