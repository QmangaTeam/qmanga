<script lang="ts">
 import HomeFixture from '../ui/HomeDiscovery.fixture.svelte';
 const params=new URLSearchParams(location.search);
 const language=params.get('locale')==='ru'?'ru':'kk';
 const size=params.get('size')==='8'?8:1;
</script>
<main><p class="fixture-note">Synthetic homepage fixture · no catalogue changes</p><HomeFixture {language} {size} topFilled distinctTop={size===8} hasContinuation={params.has('returning')} /></main>
<style>main{max-width:1280px;margin:auto;padding:16px;min-width:0}.fixture-note{font-size:12px;color:var(--muted);margin-bottom:16px}</style>
