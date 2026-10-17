<script lang="ts">
 import type {NavigationItem} from './types';
 let {items,currentPath,label}:{items:Pick<NavigationItem,'href'|'label'>[];currentPath:string;label:string}=$props();
 function active(href:string){return href==='/admin'||href==='/moderation'?currentPath===href:currentPath===href||currentPath.startsWith(href+'/');}
</script>
<nav class="workspace-nav" aria-label={label}>{#each items as item}<a href={item.href} aria-current={active(item.href)?'page':undefined}>{item.label}</a>{/each}</nav>
<style>
 .workspace-nav{display:flex;gap:4px;overflow-x:auto;border-bottom:1px solid var(--line);margin:0 0 24px;scrollbar-width:thin}
 a{display:flex;align-items:center;min-height:48px;padding:0 16px;white-space:nowrap;color:var(--muted);font-size:14px;text-decoration:none;border-bottom:2px solid transparent}
 a[aria-current]{color:var(--text);border-color:var(--accent);font-weight:650}a:focus-visible{outline:2px solid var(--accent);outline-offset:-3px}
 @media(max-width:420px){a{padding-inline:12px}}
</style>
