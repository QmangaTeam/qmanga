<script lang="ts">
 import Projects from '$lib/features/publishing/presentation/Projects.svelte';
 import { createPublishing,createTeams,createProjectRequests } from '$lib/composition/browser';
 import type { PageData } from './$types';
 const gateway = createPublishing(); let { data }: { data: PageData } = $props();
</script>
<Projects user={data.user} {gateway} teamsGateway={createTeams()} requestsGateway={createProjectRequests()}/>
