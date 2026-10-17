<script lang="ts">
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 let { title, releases, readLabel }: { title:string; releases:{id:string;teamName:string;teamSlug:string}[]; readLabel:string } = $props();
</script>
<article><h3>{title}</h3><div class="releases">{#each releases as release}<div class="release"><a class="team" href={'/teams/'+encodeURIComponent(release.teamSlug)}>{release.teamName}</a><ActionLink href={'/read/'+encodeURIComponent(release.id)}>{readLabel}</ActionLink></div>{/each}</div></article>
<style>article { padding:16px 0; border-bottom:1px solid var(--line); } h3 { margin:0 0 8px; font-size:16px; line-height:1.5; font-weight:550; overflow-wrap:anywhere; } .releases { display:grid; gap:8px; } .release { display:flex; justify-content:space-between; align-items:center; gap:16px; } .team { color:var(--muted); font-size:14px; min-height:44px; display:flex; align-items:center; overflow-wrap:anywhere; text-decoration:none; } .release :global(.ui-action) { flex-shrink:0; }</style>
