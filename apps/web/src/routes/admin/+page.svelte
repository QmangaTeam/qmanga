<script lang="ts">
 import AdminWorkQueue from '$lib/features/metrics/presentation/AdminWorkQueue.svelte';
 import Metrics from '$lib/features/metrics/presentation/Metrics.svelte';
 import {createMetrics,createWorkQueue} from '$lib/composition/browser';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {data}=$props();const {locale}=useLocale();const queue=createWorkQueue();
</script>
<AdminWorkQueue gateway={queue} actor={data.user?.id||''}/>
<details class="site-metrics"><summary>{text($locale,'Сайт статистикасы','Статистика сайта')}</summary><Metrics gateway={createMetrics()} admin/></details>
<style>.site-metrics{margin-top:28px;border-top:1px solid var(--line);padding-top:12px}.site-metrics>summary{min-height:44px;cursor:pointer;color:var(--muted);font-size:14px}</style>
