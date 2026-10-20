<script lang="ts">
 import type {ProjectListItem} from '../application/publishing';
 import ProjectCover from './ProjectCover.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {item,returnTo=''}:{item:ProjectListItem;returnTo?:string}=$props();const {locale}=useLocale();
 const href=$derived('/studio/projects/'+encodeURIComponent(item.id)+(item.nextRevisionId?'?revision='+encodeURIComponent(item.nextRevisionId):'')+(returnTo?(item.nextRevisionId?'&':'?')+'projectsReturn='+encodeURIComponent(returnTo):''));
 const work=$derived(({draft:text($locale,'Нобайды жалғастыру','Продолжить черновик'),processing:text($locale,'Өңделуде','Обработка страниц'),ready:text($locale,'Тексеруге дайын','Готово к проверке'),submitted:text($locale,'Шешімді күту','Ожидает решения'),changes_requested:text($locale,'Түзетулерді ашу','Открыть замечания')} as Record<string,string>)[item.nextRevisionState||'']);
</script>
<article class="project-card">
 <a class="cover-link" href={href} aria-label={item.titleName}><ProjectCover url={item.coverURL} name={item.titleName}/></a>
 <div class="project-info"><h2><a href={href}>{item.titleName}</a></h2><p class="team">{item.teamName}</p>
 <dl><div><dt>{text($locale,'Жоба тараулары','Глав проекта')}</dt><dd>{item.chapterCount??'—'}</dd></div><div><dt>{text($locale,'Тайтл бағасы','Оценка тайтла')}</dt><dd>{item.ratingCount&&item.score!==undefined?item.score.toFixed(2):'—'}</dd></div></dl>
 <a class="manage" href={href}>{work||text($locale,'Тарауларды басқару','Управлять главами')} <span aria-hidden="true">→</span></a></div>
</article>
<style>
 .project-card{display:flex;align-items:start;gap:14px;padding:16px;border:1px solid var(--line);border-radius:12px;background:var(--panel)}.project-info{min-width:0;flex:1}h2{font-size:17px;line-height:1.35;margin:0}h2 a{display:flex;align-items:center;min-height:44px;overflow-wrap:anywhere}.team{color:var(--muted);font-size:13px;margin:0 0 12px}dl{display:flex;flex-wrap:wrap;gap:12px 24px;margin:0}dt{font-size:12px;color:var(--muted)}dd{margin:2px 0 0;font-size:16px;font-weight:650;font-variant-numeric:tabular-nums}.manage{display:flex;align-items:center;gap:8px;min-height:44px;font-size:13px;color:var(--text);margin-top:8px}
 @media(max-width:359px){.project-card{padding:12px;gap:10px}h2{font-size:16px}dl{gap:10px}}
</style>
