<script lang="ts">
 import Catalogue from '$lib/features/catalog/presentation/Catalogue.svelte';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
<Catalogue view={data}/>
