<script lang="ts">
 import AdminTaxonomy from '$lib/features/catalog/presentation/AdminTaxonomy.svelte';
 import {createCatalogueManagement,createMetadataImports} from '$lib/composition/browser';
 import type {PageData} from './$types';
 let {data}:{data:PageData}=$props();
 const gateway=createCatalogueManagement();
</script>
<AdminTaxonomy imports={createMetadataImports()} actor={data.user?.id||''} {gateway}/>
