<script lang="ts">
 let {url,name}:{url?:string;name:string}=$props();
 let failed=$state(false);
 $effect(()=>{url;failed=false;});
</script>
<div class="project-cover" aria-hidden="true">{#if url&&!failed}<img src={url} alt="" width="80" height="120" loading="lazy" onerror={()=>failed=true}/>{:else}<span>{name.slice(0,1)}</span>{/if}</div>
<style>
 .project-cover{width:72px;aspect-ratio:2/3;flex:none;overflow:hidden;border-radius:7px;background:var(--bg);border:1px solid var(--line);display:grid;place-items:center;color:var(--muted);font-size:24px}.project-cover img{width:100%;height:100%;object-fit:cover}
 @media(min-width:600px){.project-cover{width:88px}}
</style>
