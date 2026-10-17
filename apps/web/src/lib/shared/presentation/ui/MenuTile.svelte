<script lang="ts">
 import Icon from './Icon.svelte';
 import type {IconName} from './types';
 let {href,label,icon,row=false}:{href:string;label:string;icon:IconName;row?:boolean}=$props();
</script>
<a {href} class:row><span class="menu-icon"><Icon name={icon} size={24}/></span><span class="menu-label">{label}</span>{#if row}<span class="chevron" aria-hidden="true">›</span>{/if}</a>
<style>
 a{display:flex;flex-direction:column;align-items:flex-start;justify-content:center;gap:10px;min-width:0;min-height:96px;padding:14px;border-radius:10px;border:1px solid var(--line);background:var(--panel);color:var(--text);text-decoration:none}.menu-icon{display:flex;align-items:center;justify-content:center;color:var(--accent);flex-shrink:0}.menu-label{font-size:14px;font-weight:550;line-height:1.35;overflow-wrap:anywhere;min-width:0}.row{flex-direction:row;align-items:center;justify-content:flex-start;min-height:60px;padding:10px 14px;gap:12px}.row .menu-icon{width:36px;height:36px;border-radius:8px;background:var(--raised)}.row .menu-label{flex:1}.chevron{color:var(--muted);font-size:22px}a:hover{border-color:var(--muted)}a:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
</style>
