<script lang="ts">
 import AuthForm from '$lib/features/identity/presentation/AuthForm.svelte';
 import { createAuth } from '$lib/composition/browser';
 const gateway = createAuth();
</script>
<AuthForm mode="login" {gateway}/>
