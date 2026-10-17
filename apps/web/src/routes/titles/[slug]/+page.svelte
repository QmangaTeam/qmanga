<script lang="ts">
 import Title from '$lib/features/catalog/presentation/Title.svelte';
 import LibrarySelection from '$lib/features/reading/presentation/LibrarySelection.svelte';
 import {createReading,createCommunity} from '$lib/composition/browser';
 import Discussion from '$lib/features/community/presentation/Discussion.svelte';
 const community=createCommunity();
 const reading=createReading();
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
{#key data.title.id}
<Title title={data.title} chapters={data.chapters} nextURL={data.nextURL} firstChapter={data.firstChapter} chapterFilter={data.chapterFilter}>{#snippet library()}<LibrarySelection title={data.title} readingGateway={reading.readingGateway}/>{/snippet}</Title>
<Discussion titleId={data.title.id} gateway={community}/>

{/key}
