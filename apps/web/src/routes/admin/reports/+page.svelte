<script lang="ts">
 import Reports from '$lib/features/community/presentation/Reports.svelte';
 import {createCommunity} from '$lib/composition/browser';
 const gateway=createCommunity();
</script>
<Reports {gateway}/>
