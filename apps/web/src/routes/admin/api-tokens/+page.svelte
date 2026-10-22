<script lang="ts">
 import ApiTokens from '$lib/features/identity/presentation/ApiTokens.svelte';
 import {createApiTokens} from '$lib/composition/browser';
 import type {PageData} from './$types';
 let {data}:{data:PageData}=$props();
 const gateway=createApiTokens();
</script>
<ApiTokens user={data.user} {gateway}/>
