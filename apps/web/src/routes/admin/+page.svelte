<script lang="ts">
 import TeamApprovals from '$lib/features/teams/presentation/TeamApprovals.svelte';
 import type { PageData } from './$types';
 import { createTeams } from '$lib/composition/browser';
 const gateway = createTeams();
 let { data }: { data: PageData } = $props();
</script>
<TeamApprovals user={data.user} {gateway}/>
