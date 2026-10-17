<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import type { PublicTeam } from '../application/catalog';
 let { team }: { team: PublicTeam } = $props();
</script>
<svelte:head><title>{team.name} — Qmanga</title></svelte:head>
<h1>{team.name}</h1><p style="white-space:pre-wrap;overflow-wrap:anywhere">{team.description}</p>
<a class="auth-link" href={'/updates?team='+encodeURIComponent(team.slug)}>{text($locale,'Команданың жаңа тараулары','Свежие главы команды')}</a>
{#if team.members?.length}<section><h2>{text($locale,'Команда құрамы','Состав команды')}</h2><ul>{#each team.members as member}<li class="member"><strong>{member.name}</strong><span>{member.role==='owner'?text($locale,'Иесі','Владелец'):member.role==='editor'?text($locale,'Редактор','Редактор'):text($locale,'Қатысушы','Участник')}</span></li>{/each}</ul></section>{/if}
<div class="section-heading"><h2>{copy($locale,"Жобалар")}</h2></div>
<ul>{#each team.projects as project}<li><a class="auth-link" href={'/titles/' + encodeURIComponent(project.titleSlug)}>{project.titleName}</a></li>{:else}<li>{copy($locale,"Жарияланған жоба жоқ.")}</li>{/each}</ul>
<style>section{margin-top:24px}.member{display:flex;justify-content:space-between;gap:16px;padding:12px 0;border-bottom:1px solid var(--line)}.member strong{overflow-wrap:anywhere}.member span{color:var(--muted);font-size:14px}</style>
