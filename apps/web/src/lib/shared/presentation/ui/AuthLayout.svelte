<script lang="ts">
 import type { Snippet } from 'svelte';
 import SectionHeading from './SectionHeading.svelte';
 let { title, description, children, footer }: { title:string; description?:string; children:Snippet; footer?:Snippet } = $props();
</script>
<section class="auth-layout"><a class="auth-brand" href="/" aria-label="Qmanga"><img src="/favicon.png" alt="" width="48" height="48"/><span aria-hidden="true">manga<span class="brand-dot">.</span></span></a><SectionHeading level={1} {title} {description}/>{@render children()}{#if footer}<div class="auth-footer">{@render footer()}</div>{/if}</section>
<style>.auth-brand { display:inline-flex; align-items:center; gap:2px; min-height:48px; margin-bottom:24px; color:var(--text); text-decoration:none; font-size:34px; font-weight:850; letter-spacing:-1.4px; line-height:1; } .auth-brand img { flex:none; } .brand-dot { color:var(--accent); } .auth-brand:focus-visible { outline:2px solid var(--accent); outline-offset:4px; border-radius:4px; } .auth-layout { max-width:420px; margin:8px auto 32px; } .auth-footer { margin-top:24px; padding-top:16px; border-top:1px solid var(--line); font-size:14px; color:var(--muted); } @media(min-width:768px) { .auth-layout { margin-top:32px; } }</style>
