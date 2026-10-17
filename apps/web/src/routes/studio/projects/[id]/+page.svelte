<script lang="ts">
 import PublishingStudio from '$lib/features/publishing/presentation/PublishingStudio.svelte';
 import { createPublishing } from '$lib/composition/browser';
 import { page } from '$app/stores';
 import type { PageData } from './$types';
 const gateway = createPublishing(); let { data }: { data: PageData } = $props();
</script>
{#key $page.params.id}<PublishingStudio user={data.user} projectId={$page.params.id || ''} {gateway}/>{/key}
