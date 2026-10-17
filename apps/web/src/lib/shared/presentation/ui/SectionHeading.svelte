<script lang="ts">
 import type { Snippet } from 'svelte';
 let { level=2, title, description, actions }: { level?:1|2|3; title:string; description?:string; actions?:Snippet } = $props();
</script>
<div class="section-heading">
 <div><svelte:element this={'h'+level}>{title}</svelte:element>{#if description}<p>{description}</p>{/if}</div>
 {#if actions}<div class="heading-actions">{@render actions()}</div>{/if}
</div>
<style>
 .section-heading { display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:12px; margin:0 0 24px; }
 .section-heading>div { min-width:0; } h1,h2,h3,p { margin:0; overflow-wrap:anywhere; } h1 { font-size:24px; line-height:1.3; font-weight:650; } h2 { font-size:20px; line-height:1.4; font-weight:650; } h3 { font-size:16px; line-height:1.5; font-weight:650; } p { color:var(--muted); font-size:14px; margin-top:8px; max-width:640px; } .heading-actions { display:flex; flex-wrap:wrap; gap:8px; }
 @media(min-width:768px) { h1 { font-size:28px; } }
</style>
