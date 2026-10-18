<script lang="ts">
 import Metrics from '$lib/features/metrics/presentation/Metrics.svelte';
 import {createMetrics} from '$lib/composition/browser';
</script>
<Metrics gateway={createMetrics()} admin/>
