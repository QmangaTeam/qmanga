<script lang="ts">
 let { href, name, coverURL='', metadata='', detail='', fallback='Мұқаба жоқ', priority=false }: { href:string; name:string; coverURL?:string; metadata?:string; detail?:string; fallback?:string; priority?:boolean } = $props();
</script>
<a class="cover-card" {href}><div class="cover">{#if coverURL}<img src={coverURL} alt="" width="280" height="400" loading={priority?'eager':'lazy'} fetchpriority={priority?'high':'auto'} decoding="async"/>{:else}<span>{fallback}</span>{/if}</div><h3>{name}</h3>{#if metadata}<p>{metadata}</p>{/if}{#if detail}<p>{detail}</p>{/if}</a>
<style>
 .cover-card { display:block; min-width:0; color:var(--text); text-decoration:none; } .cover { aspect-ratio:7/10; background:var(--panel); border-radius:6px; overflow:hidden; display:grid; place-items:center; color:var(--muted); font-size:13px; } img { width:100%; height:100%; object-fit:cover; } .cover span { padding:12px; text-align:center; } h3 { font-size:15px; font-weight:600; line-height:1.4; margin:8px 0 0; overflow-wrap:anywhere; } p { font-size:13px; color:var(--muted); margin:4px 0 0; } a:hover h3 { color:var(--accent); } a:focus-visible { outline:2px solid var(--accent); outline-offset:4px; border-radius:6px; }
</style>
