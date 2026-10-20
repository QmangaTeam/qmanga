<script lang="ts">
 import AdminCatalogue from '$lib/features/catalog/presentation/AdminCatalogue.svelte';
 import { createCatalogueManagement,createMetadataImports } from '$lib/composition/browser';
 import type { PageData } from './$types';
 import { page } from '$app/stores';
 const gateway = createCatalogueManagement();
 let { data }: { data: PageData } = $props();
</script>
<AdminCatalogue imports={createMetadataImports()} user={data.user} {gateway} initialTitleId={$page.url.searchParams.get('title') || ''}/>
