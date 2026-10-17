<script lang="ts">
 import Title from '$lib/features/catalog/presentation/Title.svelte';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
</script>
<Title title={data.title} chapters={data.chapters} nextURL={data.nextURL}/>
