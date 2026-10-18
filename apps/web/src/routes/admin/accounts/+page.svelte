<script lang="ts">
 import AccountManagement from '$lib/features/identity/presentation/AccountManagement.svelte';
 import {createAccountManagement} from '$lib/composition/browser';
 import type {PageData} from './$types';
 let {data}:{data:PageData}=$props();
 const gateway=createAccountManagement();
</script>
<AccountManagement user={data.user} {gateway}/>
