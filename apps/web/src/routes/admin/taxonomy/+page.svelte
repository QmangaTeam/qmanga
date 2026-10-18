<script lang="ts">
 import AdminTaxonomy from '$lib/features/catalog/presentation/AdminTaxonomy.svelte';
 import {createCatalogueManagement} from '$lib/composition/browser';
 const gateway=createCatalogueManagement();
</script>
<AdminTaxonomy {gateway}/>
