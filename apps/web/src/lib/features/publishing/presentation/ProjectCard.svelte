<script lang="ts">
 import type {ProjectListItem} from '../application/publishing';
 import ProjectCover from './ProjectCover.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {item,returnTo=''}:{item:ProjectListItem;returnTo?:string}=$props();const {locale}=useLocale();
 const revisionHref=$derived('/studio/projects/'+encodeURIComponent(item.id)+(item.nextRevisionId?'?revision='+encodeURIComponent(item.nextRevisionId):'')+(returnTo?(item.nextRevisionId?'&':'?')+'projectsReturn='+encodeURIComponent(returnTo):''));
 const href=$derived('/studio/projects/'+encodeURIComponent(item.id)+(returnTo?'?projectsReturn='+encodeURIComponent(returnTo):''));
 const work=$derived(({draft:text($locale,'Нобайды жалғастыру','Продолжить черновик'),processing:text($locale,'Өңделуде','Обработка страниц'),ready:text($locale,'Тексеруге дайын','Готово к проверке'),submitted:text($locale,'Шешімді күту','Ожидает решения'),changes_requested:text($locale,'Түзетулерді ашу','Открыть замечания')} as Record<string,string>)[item.nextRevisionState||'']);
</script>
<article class="project-card">
 <a class="cover-link" href={href} aria-label={item.titleName}><ProjectCover url={item.coverURL} name={item.titleName}/></a>
 <div class="project-info"><div class="project-identity"><h2><a href={href}>{item.titleName}</a></h2><p class="team">{item.teamName}</p></div>
 <dl><div><dt>{text($locale,'Жоба тараулары','Глав проекта')}</dt><dd>{item.chapterCount??'—'}</dd></div><div><dt>{text($locale,'Тайтл бағасы','Оценка тайтла')}</dt><dd>{item.ratingCount&&item.score!==undefined?item.score.toFixed(2):'—'}</dd></div></dl>
 <div class="project-actions"><a class="manage" href={href}>{text($locale,'Жобаны ашу','Открыть проект')} <span aria-hidden="true">→</span></a>{#if item.nextRevisionId&&work}<a class="resume" href={revisionHref}>{work}</a>{/if}</div></div>
</article>
<style>
 .project-card{display:flex;align-items:center;gap:14px;padding:12px 10px;border-bottom:1px solid var(--line);min-width:0}.project-card:hover{background:var(--panel)}.cover-link{width:42px;flex:none}.cover-link :global(.project-cover){width:42px;height:60px}.project-info{display:grid;grid-template-columns:minmax(0,1fr) 150px 180px;align-items:center;gap:20px;min-width:0;flex:1}.project-identity{min-width:0}h2{font-size:14px;line-height:1.4;font-weight:600;margin:0}h2 a{display:flex;align-items:center;min-height:28px;overflow-wrap:anywhere}.team{color:var(--muted);font-size:12px;margin:3px 0 0;overflow-wrap:anywhere}dl{display:flex;gap:16px;margin:0}dt{font-size:10px;color:var(--muted);font-weight:400}dd{margin:4px 0 0;font-size:13px;font-variant-numeric:tabular-nums}.manage{display:flex;align-items:center;justify-content:space-between;gap:8px;min-height:44px;font-size:12px;color:var(--accent)}
 @media(max-width:899px){.project-info{grid-template-columns:minmax(0,1fr) 140px;gap:4px 16px}.project-identity{grid-column:1/-1}dl{gap:16px}.manage{justify-self:end;max-width:160px}}
 @media(max-width:479px){.project-card{align-items:start;padding-inline:0;gap:10px}.project-info{grid-template-columns:minmax(0,1fr);gap:8px}.project-identity{grid-column:auto}.manage{justify-self:start;min-height:44px;max-width:none}h2 a{min-height:44px}dl{gap:24px}}
.project-actions{min-width:0}.resume{display:flex;align-items:center;min-height:44px;font-size:12px;color:var(--muted);overflow-wrap:anywhere}.project-actions .manage{justify-self:stretch}
</style>
