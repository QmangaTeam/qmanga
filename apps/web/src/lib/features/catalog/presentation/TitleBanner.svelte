<script lang="ts">
 let {bannerURL='',coverURL=''}:{bannerURL?:string;coverURL?:string}=$props();
 let failed=$state<string[]>([]);
 const source=$derived([bannerURL,coverURL].find(url=>url&&!failed.includes(url))||'');
</script>
{#if source}<div class="title-banner" class:fallback={source!==bannerURL} aria-hidden="true"><img src={source} alt="" width="1280" height="320" onerror={()=>failed=[...failed,source]}/></div>{/if}
<style>.title-banner{height:clamp(96px,14vw,180px);width:100%;overflow:hidden;border-radius:12px;margin:12px 0 20px;background:var(--panel)}img{display:block;width:100%;height:100%;object-fit:cover;object-position:center 35%}@media(max-width:767px){.title-banner{height:96px;border-radius:8px;margin-bottom:16px}.title-banner.fallback{display:none}}</style>
