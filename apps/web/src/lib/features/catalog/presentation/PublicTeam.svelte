<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import type { PublicTeam } from '../application/catalog';
 let { team }: { team: PublicTeam } = $props();
 const initials=$derived(team.name.trim().split(/\s+/).slice(0,2).map(word=>[...word][0]||'').join('').toLocaleUpperCase());
</script>
<svelte:head><title>{team.name} — Qmanga</title></svelte:head>
<div class="team-header"><span class="avatar" aria-hidden="true">{initials}</span><div><h1>{team.name}</h1>{#if team.members?.length===1}<p class="muted">{text($locale,'Жеке аудармашы','Один переводчик')}</p>{/if}</div></div><p class="description">{team.description}</p>
<div class="actions"><a class="button" href={'/updates?team='+encodeURIComponent(team.slug)}>{text($locale,'Команданың жаңа тараулары','Свежие главы команды')}</a>{#if team.telegramURL}<a class="button" href={team.telegramURL} target="_blank" rel="noopener noreferrer">Telegram ↗</a>{/if}</div>
{#if team.members?.length}<section><h2>{text($locale,'Команда құрамы','Состав команды')}</h2><ul>{#each team.members as member}<li class="member"><strong>{member.name}</strong><span>{member.role==='owner'?text($locale,'Иесі','Владелец'):member.role==='editor'?text($locale,'Редактор','Редактор'):text($locale,'Қатысушы','Участник')}</span></li>{/each}</ul></section>{/if}
<div class="section-heading"><h2>{copy($locale,"Жобалар")}</h2></div>
<ul>{#each team.projects as project}<li><a class="auth-link" href={'/titles/' + encodeURIComponent(project.titleSlug)}>{project.titleName}</a></li>{:else}<li>{copy($locale,"Жарияланған жоба жоқ.")}</li>{/each}</ul>
<style>.team-header{display:flex;gap:16px;align-items:center}.team-header>div{min-width:0}.team-header h1{overflow-wrap:anywhere}.avatar{display:grid;place-items:center;flex:0 0 64px;height:64px;border-radius:12px;background:var(--panel);border:1px solid var(--line);font-size:22px;font-weight:700}.description{white-space:pre-wrap;overflow-wrap:anywhere;margin:16px 0;line-height:1.6}.actions{display:flex;flex-wrap:wrap;gap:8px}.muted{color:var(--muted)}section{margin-top:24px}.member{display:flex;justify-content:space-between;gap:16px;padding:12px 0;border-bottom:1px solid var(--line)}.member strong{overflow-wrap:anywhere}.member span{color:var(--muted);font-size:14px}</style>
