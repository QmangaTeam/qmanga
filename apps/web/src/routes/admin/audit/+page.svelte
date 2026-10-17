<script lang="ts">
 import AuditLog from '$lib/features/identity/presentation/AuditLog.svelte';
 import {createAccountManagement} from '$lib/composition/browser';
 import type {PageData} from './$types';
 let {data}:{data:PageData}=$props();
 const gateway=createAccountManagement();
</script>
<AuditLog user={data.user} {gateway}/>
