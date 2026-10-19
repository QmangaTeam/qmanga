<script lang="ts">
 let {bannerURL='',coverURL=''}:{bannerURL?:string;coverURL?:string}=$props();
 let failed=$state<string[]>([]);
 const source=$derived([bannerURL,coverURL].find(url=>url&&!failed.includes(url))||'');
</script>
{#if source}<div class="title-banner" aria-hidden="true"><img src={source} alt="" width="1280" height="320" onerror={()=>failed=[...failed,source]}/></div>{/if}
<style>.title-banner{height:clamp(140px,25vw,320px);width:100%;overflow:hidden;border-radius:12px;margin:12px 0 20px;background:var(--panel)}img{display:block;width:100%;height:100%;object-fit:cover;object-position:center 35%}@media(max-width:599px){.title-banner{height:140px;border-radius:8px;margin-bottom:16px}}</style>
