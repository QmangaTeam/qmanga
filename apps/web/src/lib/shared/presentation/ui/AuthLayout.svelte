<script lang="ts">
 import type { Snippet } from 'svelte';
 import SectionHeading from './SectionHeading.svelte';
 let { title, description, children, footer }: { title:string; description?:string; children:Snippet; footer?:Snippet } = $props();
</script>
<section class="auth-layout"><SectionHeading level={1} {title} {description}/>{@render children()}{#if footer}<div class="auth-footer">{@render footer()}</div>{/if}</section>
<style>.auth-layout { max-width:420px; margin:8px auto 32px; } .auth-footer { margin-top:24px; padding-top:16px; border-top:1px solid var(--line); font-size:14px; color:var(--muted); } @media(min-width:768px) { .auth-layout { margin-top:32px; } }</style>
