<script lang="ts">
 import Studio from '$lib/features/teams/presentation/Studio.svelte';
 import {createTeams,createAvatars} from '$lib/composition/browser';
 let {data}=$props();const gateway=createTeams();
</script>
<Studio avatars={createAvatars()} user={data.user} {gateway}/>
