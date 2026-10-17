<script lang="ts">
 import { page } from '$app/stores';
 import AdminProject from '$lib/features/catalog/presentation/AdminProject.svelte';
 import { createCatalogueManagement } from '$lib/composition/browser';
 import type { PageData } from './$types';
 const gateway = createCatalogueManagement();
 let { data }: { data: PageData } = $props();
</script>
{#key $page.params.id}<AdminProject user={data.user} projectId={$page.params.id || ''} {gateway}/>{/key}
