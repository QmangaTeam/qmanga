<script lang="ts">
 import AdminCatalogue from '$lib/features/catalog/presentation/AdminCatalogue.svelte';
 import { createCatalogueManagement } from '$lib/composition/browser';
 import type { PageData } from './$types';
 const gateway = createCatalogueManagement();
 let { data }: { data: PageData } = $props();
</script>
<AdminCatalogue user={data.user} {gateway}/>
