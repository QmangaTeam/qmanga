<script lang="ts">
 import MyReports from '$lib/features/community/presentation/MyReports.svelte';
 import {createCommunity} from '$lib/composition/browser';
 const gateway=createCommunity();
</script>
<MyReports {gateway}/>
