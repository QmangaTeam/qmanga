<script lang="ts">
 import {createReading} from '$lib/composition/browser';
 import CoverStatuses from '$lib/features/reading/presentation/CoverStatuses.svelte';
 const reading=createReading();
 import Catalogue from '$lib/features/catalog/presentation/Catalogue.svelte';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
<CoverStatuses {...reading} titleIds={data.page.items.map(title=>title.id)}>{#snippet children(statuses)}<Catalogue view={data} {statuses}/>{/snippet}</CoverStatuses>
