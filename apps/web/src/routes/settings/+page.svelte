<script lang="ts">
 import Settings from '$lib/features/preferences/presentation/Settings.svelte';
 import { createAuth, createPreferences } from '$lib/composition/browser';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
 const auth = createAuth();
 const preferences = createPreferences();
</script>
<Settings user={data.user} {auth} {preferences}/>
