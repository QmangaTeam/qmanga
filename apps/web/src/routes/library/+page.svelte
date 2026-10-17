<script lang="ts">
 import Library from '$lib/features/reading/presentation/Library.svelte';
 import {createReading} from '$lib/composition/browser';
 const ports=createReading();
</script>
<Library {...ports}/>
