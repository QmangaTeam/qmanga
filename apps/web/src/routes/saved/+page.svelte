<script lang="ts">
 import Saved from '$lib/features/offline/presentation/Saved.svelte';
 import {createDownloads} from '$lib/composition/browser';
 const downloads=createDownloads();
</script>
<Saved {downloads}/>
