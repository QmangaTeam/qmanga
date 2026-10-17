<script lang="ts">
 import type { Snippet } from 'svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 let { name, href, detail, releaseId, continueLabel, children }: {name:string;href:string;detail?:string;releaseId?:string;continueLabel:string;children?:Snippet} = $props();
</script>
<article><div class="row-heading"><div><a class="title" {href}>{name}</a>{#if detail}<p>{detail}</p>{/if}</div>{#if releaseId}<ActionLink href={'/read/'+encodeURIComponent(releaseId)} variant="primary">{continueLabel}</ActionLink>{/if}</div>{#if children}<div class="row-controls">{@render children()}</div>{/if}</article>
<style>article { border-bottom:1px solid var(--line); padding:16px 0; } .row-heading { display:flex; flex-wrap:wrap; align-items:center; justify-content:space-between; gap:12px; } .row-heading>div { min-width:0; } .title { min-height:44px; display:inline-flex; align-items:center; font-size:16px; font-weight:600; color:var(--text); text-decoration:none; overflow-wrap:anywhere; } p { color:var(--muted); font-size:14px; margin:0; } .row-controls { display:flex; flex-wrap:wrap; align-items:center; gap:12px; margin-top:12px; }</style>
