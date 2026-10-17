<script lang="ts">
 import {page} from '$app/stores';
 import {onMount} from 'svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import OfflineReader from '$lib/features/offline/presentation/OfflineReader.svelte';
 import {createDownloads,createOfflineReader} from '$lib/composition/browser';
 const downloads=createDownloads(),ports=createOfflineReader();
 const {locale}=useLocale();
 let ready=$state(false),releaseId=$state<string|undefined>();
 onMount(()=>{releaseId=$page.url.searchParams.get('release')||undefined;ready=true;});
</script>
{#if ready}<OfflineReader {downloads} {ports} {releaseId}/>{:else}<p class="page-container" role="status">{text($locale,'Сақталған тарау тексерілуде…','Проверяем скачанную главу…')}</p>{/if}
