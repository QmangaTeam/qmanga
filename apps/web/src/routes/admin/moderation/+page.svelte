<script lang="ts">
 import Moderation from '$lib/features/publishing/presentation/Moderation.svelte';
 import { createPublishing,createReader } from '$lib/composition/browser';
 import type { PageData } from './$types';
 const gateway = createPublishing(); let { data }: { data: PageData } = $props();
</script>
<Moderation user={data.user} {gateway} reader={createReader().gateway}/>
