<script lang="ts">import {page} from '$app/stores';import TitleChangeDetail from '$lib/features/catalog/presentation/TitleChangeDetail.svelte';import {createTitleChanges} from '$lib/composition/browser';const gateway=createTitleChanges();</script>
{#key $page.params.id}<TitleChangeDetail {gateway} id={$page.params.id||''}/>{/key}
