<script lang="ts">
 import {page} from '$app/stores';
 import TitleChangeForm from '$lib/features/catalog/presentation/TitleChangeForm.svelte';
 import {createTitleChanges} from '$lib/composition/browser';
 const gateway=createTitleChanges();
</script>
{#key $page.params.id}<TitleChangeForm {gateway} projectId={$page.params.id||''} coverPageId={$page.url.searchParams.get('coverPage')||''}/>{/key}
