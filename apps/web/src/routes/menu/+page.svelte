<script lang="ts">
 import Menu from '$lib/features/preferences/presentation/Menu.svelte';
 let {data}=$props();
</script>
<Menu user={data.user}/>
