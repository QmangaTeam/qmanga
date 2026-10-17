<script lang="ts">
 import Studio from '$lib/features/teams/presentation/Studio.svelte';
 import type { PageData } from './$types';
 import { createTeams } from '$lib/composition/browser';
 const gateway = createTeams();
 let { data }: { data: PageData } = $props();
</script>
<Studio user={data.user} {gateway}/>
