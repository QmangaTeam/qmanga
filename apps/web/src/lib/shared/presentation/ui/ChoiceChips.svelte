<script lang="ts">
 import Icon from './Icon.svelte';
 import type {IconName} from './types';
 let {label,options,value,onchoose,disabled=false}:{label:string;options:{value:string;label:string;icon?:IconName}[];value:string;onchoose:(value:string)=>void;disabled?:boolean}=$props();
</script>
<div class="choice-chips" role="group" aria-label={label}>
 {#each options as option}<button type="button" data-value={option.value} aria-pressed={value===option.value} {disabled} onclick={()=>onchoose(option.value)}>{#if option.icon}<Icon name={option.icon} size={18}/>{/if}<span>{option.label}</span></button>{/each}
</div>
<style>
 .choice-chips{display:flex;gap:4px;padding:4px;border:1px solid var(--line);border-radius:10px;background:var(--bg);min-width:0;max-width:100%}button{display:flex;align-items:center;justify-content:center;flex-direction:column;gap:6px;flex:1 1 0;min-width:44px;min-height:44px;padding:8px;border:1px solid transparent;border-radius:7px;background:transparent;color:var(--muted);font-size:14px;font-weight:550;line-height:1.3;overflow-wrap:anywhere}button[aria-pressed=true]{color:var(--accent);background:var(--raised);border-color:var(--line)}button:not(:disabled):hover{color:var(--text)}button:focus-visible{outline:2px solid var(--accent);outline-offset:1px}span{min-width:0;white-space:nowrap}
</style>
