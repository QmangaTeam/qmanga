<script lang="ts">
 import Team from '$lib/features/catalog/presentation/PublicTeam.svelte';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
<Team team={data.team}/>
