<script lang="ts">
 import Studio from '$lib/features/teams/presentation/Studio.svelte';
 import {createTeams} from '$lib/composition/browser';
 let {data}=$props();const gateway=createTeams();
</script>
<Studio user={data.user} {gateway}/>
