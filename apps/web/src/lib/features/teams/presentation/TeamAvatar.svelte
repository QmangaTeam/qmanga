<script lang="ts">
 let {id,name,size=56,version=0}:{id:string;name:string;size?:number;version?:number}=$props();let failed=$state(false);
 $effect(()=>{id;version;failed=false;});
</script>
<span class="team-avatar" style:--avatar-size={size+'px'} aria-hidden="true"><span>{name.trim().slice(0,1).toLocaleUpperCase()||'Q'}</span>{#if !failed}<img src={'/media/avatars/team/'+encodeURIComponent(id)+(version?'?v='+version:'')} alt="" width={size} height={size} loading="lazy" onerror={()=>failed=true}/>{/if}</span>
<style>.team-avatar{position:relative;display:grid;place-items:center;width:var(--avatar-size);height:var(--avatar-size);flex:none;overflow:hidden;border:1px solid var(--line);border-radius:12px;background:var(--raised);color:var(--muted);font-size:22px;font-weight:600}.team-avatar img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover}</style>
