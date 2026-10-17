<script lang="ts">
 import type { Snippet } from 'svelte';
 let { controlId, label, description, children }: { controlId:string; label:string; description?:string; children:Snippet } = $props();
</script>
<div class="settings-row"><div><label for={controlId}>{label}</label>{#if description}<p>{description}</p>{/if}</div><div class="setting-control">{@render children()}</div></div>
<style>
 .settings-row { display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:12px 24px; padding:16px 0; border-bottom:1px solid var(--line); } .settings-row>div { min-width:0; } label { font-weight:550; } p { color:var(--muted); font-size:14px; max-width:480px; margin:4px 0 0; } .setting-control { flex:1 1 160px; max-width:100%; } .setting-control :global(select) { width:100%; min-height:44px; font-size:16px; background-color:var(--panel); color:var(--text); border:1px solid var(--muted); border-radius:8px; padding:10px 40px 10px 12px; } @media(min-width:640px) { .setting-control { flex:0 1 240px; } }
</style>
