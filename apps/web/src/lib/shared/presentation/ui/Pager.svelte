<script lang="ts">
 import Button from './Button.svelte';
 import ActionLink from './ActionLink.svelte';
 import type { PageControl } from './types';
 let { previous, next, label }: { previous?:PageControl; next?:PageControl; label:string } = $props();
</script>
<nav aria-label={label}>
 {#each [previous,next] as control}{#if control}{#if 'href' in control}<ActionLink href={control.href}>{control.label}</ActionLink>{:else}<Button onclick={control.action} busy={control.busy} disabled={control.disabled}>{control.label}</Button>{/if}{/if}{/each}
</nav>
<style>nav { display:flex; flex-wrap:wrap; justify-content:space-between; gap:12px; margin-top:24px; }</style>
