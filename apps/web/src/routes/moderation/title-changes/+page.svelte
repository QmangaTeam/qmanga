<script lang="ts">import TitleChangeQueue from '$lib/features/catalog/presentation/TitleChangeQueue.svelte';import {createTitleChanges} from '$lib/composition/browser';const gateway=createTitleChanges();</script>
<TitleChangeQueue {gateway}/>
