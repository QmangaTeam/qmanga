<script lang="ts">
 import AdminQueues from '$lib/features/metrics/presentation/AdminQueues.svelte';
 import Metrics from '$lib/features/metrics/presentation/Metrics.svelte';
 import {createMetrics} from '$lib/composition/browser';
</script>
<AdminQueues/>
<Metrics gateway={createMetrics()} admin/>
