<script lang="ts">
 import {createReading} from '$lib/composition/browser';
 import ContinueReading from '$lib/features/reading/presentation/ContinueReading.svelte';
 import HomeDiscovery from '$lib/features/catalog/presentation/HomeDiscovery.svelte';
 import type {PageData} from './$types';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 const reading=createReading();
 let {data}:{data:PageData}=$props();
 const {locale}=useLocale();
</script>
<svelte:head><title>{text($locale,'Qmanga — қазақша манга','Qmanga — манга на казахском')}</title></svelte:head>
<HomeDiscovery mostRead={data.mostRead} updates={data.updates} top={data.top} topKind={data.topKind}>
 {#snippet continuation()}<ContinueReading {...reading} covers={data.mostRead.page.items}/>{/snippet}
</HomeDiscovery>
