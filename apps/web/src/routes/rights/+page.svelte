<script lang="ts">
 import {page} from '$app/state';
 import Rights from '$lib/features/community/presentation/Rights.svelte';
 import {createRights} from '$lib/composition/browser';
 const gateway=createRights();
</script>
<Rights {gateway} initialTitle={page.url.searchParams.get('title')||''}/>
