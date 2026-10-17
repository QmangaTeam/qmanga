<script lang="ts">
 import type { Snippet } from 'svelte';
 let { kind, title, description, actions }: { kind:'empty'|'loading'|'error'|'offline'|'success'; title:string; description?:string; actions?:Snippet } = $props();
 const role = $derived(kind === 'error' ? 'alert' : kind === 'empty' ? undefined : 'status');
</script>
<div class="state-message" class:error={kind === 'error'} {role} aria-busy={kind === 'loading' ? true : undefined}>
 <p class="state-title">{title}</p>{#if description}<p class="state-description">{description}</p>{/if}
 {#if actions}<div class="state-actions">{@render actions()}</div>{/if}
</div>
<style>
 .state-message { padding:24px 0; border-block:1px solid var(--line); } .state-title { font-size:16px; font-weight:600; margin:0; overflow-wrap:anywhere; } .state-description { font-size:14px; color:var(--muted); max-width:560px; margin:8px 0 0; overflow-wrap:anywhere; } .state-actions { display:flex; flex-wrap:wrap; gap:8px; margin-top:16px; } .error { border-color:var(--muted); }
</style>
