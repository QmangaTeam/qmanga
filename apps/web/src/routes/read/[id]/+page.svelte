<script lang="ts">
 import Reader from '$lib/features/reading/presentation/Reader.svelte';
 import { createDownloads,createReader,createReading,createCommunity } from '$lib/composition/browser';
 import type { PageData } from './$types';
 let {data}:{data:PageData}=$props();
 const ports={...createReader(),...createReading()},downloads=createDownloads(),community=createCommunity();
</script>
{#key data.manifest.revisionId}<Reader manifest={data.manifest} preview={data.preview} {downloads} {community} {...ports}/>{/key}
