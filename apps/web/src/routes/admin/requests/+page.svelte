<script lang="ts">
 import ProjectRequests from '$lib/features/publishing/presentation/ProjectRequests.svelte';
 import {createProjectRequests,createTeams} from '$lib/composition/browser';
 let {data}=$props();const gateway=createProjectRequests(),teamsGateway=createTeams();
</script>
<ProjectRequests user={data.user} {gateway} {teamsGateway} admin/>
