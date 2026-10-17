<script lang="ts">
 import type { PublicTeam } from '../application/catalog';
 let { team }: { team: PublicTeam } = $props();
</script>
<svelte:head><title>{team.name} — Qmanga</title></svelte:head>
<h1>{team.name}</h1><p style="white-space:pre-wrap;overflow-wrap:anywhere">{team.description}</p>
<div class="section-heading"><h2>Жобалар</h2></div>
<ul>{#each team.projects as project}<li><a class="auth-link" href={'/titles/' + encodeURIComponent(project.titleSlug)}>{project.titleName}</a></li>{:else}<li>Жарияланған жоба жоқ.</li>{/each}</ul>
