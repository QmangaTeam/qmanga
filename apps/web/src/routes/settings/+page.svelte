<script lang="ts">
 import Settings from '$lib/features/preferences/presentation/Settings.svelte';
 import { createAccountProfiles,createReader,createAvatars,createAuth,createGoogleLink, createPreferences } from '$lib/composition/browser';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
 const auth = createAuth();
 const preferences = createPreferences();
</script>
{#key data.user?.id}<Settings accountProfiles={createAccountProfiles()} reading={createReader().platform} avatars={createAvatars()} googleLink={createGoogleLink()} user={data.user} {auth} {preferences}/>{/key}
