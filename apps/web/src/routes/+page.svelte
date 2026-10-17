<script lang="ts">
  import type { PageData } from './$types';
  let { data }: { data: PageData } = $props();
</script>
<svelte:head><title>Qmanga — Қазақша оқы</title></svelte:head>
<section class="hero">
  <span class="eyebrow">Оқиға өз тіліңде</span>
  <h1>Жаңа әлемдерді<br/><em>қазақша аш.</em></h1>
  <p>Сүйікті мангаң, манхваң және жаңа оқиғалар. Жолда да, ауылда да — өз ырғағыңмен оқы.</p>
  <a class="button primary" href="/catalog">Каталогты ашу <span aria-hidden="true">↗</span></a>
  <div class="hero-art" aria-hidden="true"><div class="orbit"></div><span>Q</span><div class="art-caption">ҚАЗАҚША<br/>ОҚИҒАЛАР</div></div>
</section>
<section class="section-heading"><div><span class="eyebrow">Жаңа тараулар</span><h2>Жаңадан қосылған</h2></div><a href="/catalog">Барлығы ↗</a></section>
<div class="empty-state"><div class="empty-symbol" aria-hidden="true">↗</div><h3>Алғашқы оқиғалар жақында</h3><p>Қазақша аудармалар осында жиналады. Оқу үшін тіркелу міндетті емес.</p></div>
<div class="connection-state" role="status"><span class:connected={data.status.available}></span>{data.status.available ? 'Оқуға дайын' : 'Қосылым қалпына келуде'}</div>
