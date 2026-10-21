<script lang="ts">
 import TitleFixture from '../ui/Title.fixture.svelte';
 const params=new URLSearchParams(location.search);
 const genreCount=Math.min(60,Math.max(0,Number(params.get('tags')||30)));
 const editionCount=Math.min(50,Math.max(1,Number(params.get('editions')||10)));
</script>
<TitleFixture {genreCount} {editionCount} language={params.get('locale')==='kk'?'kk':'ru'} tab={params.get('section')||'chapters'}/>
