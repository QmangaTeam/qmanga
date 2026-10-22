<script lang="ts">
 import {page} from '$app/state';
 import {createAccountProfiles} from '$lib/composition/browser';
 import Profile from '$lib/features/preferences/presentation/Profile.svelte';
 let {data}=$props();
 const gateway=createAccountProfiles();
</script>
<Profile id={page.params.id||''} user={data.user} {gateway}/>
