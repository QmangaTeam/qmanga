<script lang="ts">
 import {onMount} from 'svelte';
 import {avatarChanges} from '$lib/features/identity/infrastructure/avatar-changes';
 import Menu from '$lib/features/preferences/presentation/Menu.svelte';
 import {createAuth,createTeams} from '$lib/composition/browser';
 import {goto,invalidateAll} from '$app/navigation';
 let {data}=$props();
 let avatarVersion=$state(0);
 onMount(()=>avatarChanges.subscribe(actor=>{if(actor===data.user?.id)avatarVersion++;}));
 const auth=createAuth(),teams=createTeams();
 async function loggedOut(){await invalidateAll();await goto('/');}
</script>
{#key [data.user?.id,data.user?.verified,data.user?.avatarURL].join(':')}<Menu {avatarVersion} user={data.user} {auth} {teams} onlogout={loggedOut}/>{/key}
