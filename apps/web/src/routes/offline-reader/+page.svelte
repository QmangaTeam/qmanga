<script lang="ts">
 import OfflineReader from '$lib/features/offline/presentation/OfflineReader.svelte';
 import {createDownloads,createOfflineReader} from '$lib/composition/browser';
 const downloads=createDownloads(),ports=createOfflineReader();
</script>
<OfflineReader {downloads} {ports}/>
