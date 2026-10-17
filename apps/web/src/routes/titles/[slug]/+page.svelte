<script lang="ts">
 import Title from '$lib/features/catalog/presentation/Title.svelte';
 import LibrarySelection from '$lib/features/reading/presentation/LibrarySelection.svelte';
 import {createReading,createCommunity} from '$lib/composition/browser';
 import Discussion from '$lib/features/community/presentation/Discussion.svelte';
 import ReportForm from '$lib/features/community/presentation/ReportForm.svelte';
 const community=createCommunity();
 const reading=createReading();
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
{#key data.title.id}
<Title title={data.title} chapters={data.chapters} nextURL={data.nextURL} firstChapter={data.firstChapter} chapterFilter={data.chapterFilter} tab={data.tab}>
 {#snippet library()}<LibrarySelection title={data.title} readingGateway={reading.readingGateway} librarySync={reading.librarySync}/>{/snippet}
 {#snippet discussion()}<Discussion titleId={data.title.id} titleSlug={data.title.slug} gateway={community} showReport={false}/>{/snippet}
 {#snippet report()}<ReportForm titleId={data.title.id} targetId={data.title.id} gateway={community}/>{/snippet}
</Title>

{/key}
