<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {projectId,current='overview',returnTo=''}:{projectId:string;current?:string;returnTo?:string}=$props();
 const {locale}=useLocale();
 const base=$derived('/studio/projects/'+encodeURIComponent(projectId));
 function href(section:string){return base+(section==='metadata'?'/metadata':'')+'?'+(section!=='metadata'?'section='+section+'&':'')+(returnTo?'projectsReturn='+encodeURIComponent(returnTo):'');}
</script>
<nav class="project-navigation" aria-label={text($locale,'Жоба бөлімдері','Разделы проекта')}>
 {#each [['overview','Негізгі','Основное'],['chapters','Тараулар','Главы'],['metadata','Тайтл туралы','Сведения о тайтле'],['requests','Өзгерту өтінімдері','Заявки на правки']] as [id,kk,ru]}
 <a href={href(id)} aria-current={current===id?'page':undefined}>{text($locale,kk,ru)}</a>
 {/each}
</nav>
<style>
 .project-navigation{display:flex;gap:6px;flex-wrap:wrap;border-bottom:1px solid var(--line);padding-bottom:10px;margin:16px 0 24px}.project-navigation a{display:flex;align-items:center;min-height:44px;padding:8px 12px;font-size:13px;border-radius:7px;color:var(--muted);text-decoration:none}.project-navigation a[aria-current]{color:var(--text);background:var(--raised)}
 @media(max-width:479px){.project-navigation{display:grid;grid-template-columns:1fr 1fr;gap:4px}.project-navigation a{font-size:12px;overflow-wrap:anywhere}}
</style>
