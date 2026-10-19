<script lang="ts">
 import {createReading} from '$lib/composition/browser';
 import CoverStatuses from '$lib/features/reading/presentation/CoverStatuses.svelte';
 import ContinueReading from '$lib/features/reading/presentation/ContinueReading.svelte';
 import HomeDiscovery from '$lib/features/catalog/presentation/HomeDiscovery.svelte';
 import type {PageData} from './$types';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 const reading=createReading();
 let {data}:{data:PageData}=$props();
 const {locale}=useLocale();
</script>
<svelte:head><title>{text($locale,'Qmanga — қазақша манга','Qmanga — манга на казахском')}</title></svelte:head>
<CoverStatuses {...reading} titleIds={[...data.mostRead.page.items,...data.top.page.items].map(title=>title.id)}>{#snippet children(statuses)}<HomeDiscovery {statuses} mostRead={data.mostRead} updates={data.updates} top={data.top} topKind={data.topKind}>
 {#snippet continuation()}<ContinueReading {...reading}/>{/snippet}
</HomeDiscovery>{/snippet}</CoverStatuses>
