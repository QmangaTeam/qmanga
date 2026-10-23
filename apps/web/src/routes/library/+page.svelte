<script lang="ts">
 import Library from '$lib/features/reading/presentation/Library.svelte';
 import {createLibrary} from '$lib/composition/browser';
 const ports=createLibrary();
</script>
<Library {...ports}/>
