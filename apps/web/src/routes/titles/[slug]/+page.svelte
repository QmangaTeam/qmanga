<script lang="ts">
 import {useLocale} from '$lib/shared/presentation/locale';
 import {copy,type UILabel} from '$lib/shared/presentation/copy';
 import Title from '$lib/features/catalog/presentation/Title.svelte';
 import LibrarySelection from '$lib/features/reading/presentation/LibrarySelection.svelte';
 import {createReading,createCommunity} from '$lib/composition/browser';
 import Discussion from '$lib/features/community/presentation/Discussion.svelte';
 import RatingDialog from '$lib/features/community/presentation/RatingDialog.svelte';
 import ReportForm from '$lib/features/community/presentation/ReportForm.svelte';
 const {locale}=useLocale();
 let libraryStatus=$state('');
 const statusCopy:Record<string,UILabel>={reading:'Оқып жүрмін',planned:'Жоспарда',completed:'Оқып болдым',paused:'Үзілісте',dropped:'Тоқтаттым'};
 const libraryLabel=$derived(libraryStatus&&statusCopy[libraryStatus]?copy($locale,statusCopy[libraryStatus]):'');
 const community=createCommunity();
 const reading=createReading();
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
{#key data.title.id}
<Title {libraryLabel} title={data.title} chapters={data.chapters} nextURL={data.nextURL} firstChapter={data.firstChapter} chapterFilter={data.chapterFilter} tab={data.tab}>
 {#snippet library()}<LibrarySelection onstatuschange={status=>libraryStatus=status||''} title={data.title} readingGateway={reading.readingGateway} librarySync={reading.librarySync}/>{/snippet}
 {#snippet discussion()}<Discussion titleId={data.title.id} titleSlug={data.title.slug} gateway={community} showReport={false}/>{/snippet}
 {#snippet rating()}<RatingDialog titleId={data.title.id} titleSlug={data.title.slug} score={data.title.score} ratingCount={data.title.ratingCount} gateway={community}/>{/snippet}
 {#snippet report()}<ReportForm titleId={data.title.id} targetId={data.title.id} gateway={community} embedded/>{/snippet}
</Title>

{/key}
