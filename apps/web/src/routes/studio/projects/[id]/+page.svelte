<script lang="ts">
 import ProjectWorkspace from '$lib/features/publishing/presentation/ProjectWorkspace.svelte';
 import PublishingStudio from '$lib/features/publishing/presentation/PublishingStudio.svelte';
 import { createPublishing,createTitleChanges } from '$lib/composition/browser';
 import { page } from '$app/stores';
 import type { PageData } from './$types';
 const gateway = createPublishing(); let { data }: { data: PageData } = $props();
</script>
{#key $page.params.id+':'+(data.user?.id||'')}
 {#if $page.url.searchParams.has('revision')}<PublishingStudio user={data.user} projectId={$page.params.id || ''} {gateway}/>{:else}<ProjectWorkspace projectId={$page.params.id||''} {gateway} changes={createTitleChanges()}/>{/if}
{/key}
