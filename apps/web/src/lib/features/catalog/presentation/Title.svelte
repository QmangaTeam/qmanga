<script lang="ts">
 import type { TitleDetail, Chapter, ResultPage } from '../application/catalog';
 import { kindNames, originalStatusNames, translationStatusNames } from '../domain/catalog';
 let { title, chapters, nextURL }: { title: TitleDetail; chapters: ResultPage<Chapter>; nextURL: string | null } = $props();
</script>
<svelte:head><title>{title.name} — Qmanga</title><meta name="description" content={title.description.slice(0, 160)}/></svelte:head>
<a class="auth-link" href="/catalog">Каталогқа оралу</a>
<div class="title-header">{#if title.coverURL}<img class="title-cover" src={title.coverURL} alt={title.name + ' мұқабасы'} width="280" height="400" fetchpriority="high"/>{/if}<div><h1>{title.name}</h1>{#if title.originalName}<p class="muted">{title.originalName}</p>{/if}<p>{kindNames[title.kind]}{#if title.demo} / Үлгі туынды{/if}</p><dl><dt>Түпнұсқа</dt><dd>{originalStatusNames[title.originalStatus]}</dd><dt>Аударма</dt><dd>{translationStatusNames[title.translationStatus]}</dd></dl>
{#if title.ratingCount > 0}<p>{title.score.toFixed(1)} / 10 ({title.ratingCount} баға)</p>{/if}
<div class="title-genres">{#each title.genres as genre}<a class="button subtle" href={'/catalog?genre=' + encodeURIComponent(genre.slug)}>{genre.name}</a>{/each}</div></div></div>
{#if title.description}<section class="title-section"><h2>Сипаттама</h2><p class="description">{title.description}</p></section>{/if}
{#if title.aliases.length}<details><summary>Басқа атаулары</summary><ul>{#each title.aliases as alias}<li>{alias}</li>{/each}</ul></details>{/if}
<section class="title-section"><h2>Аударма командалары</h2><ul>{#each title.projects as project}<li><a class="auth-link" href={'/teams/' + encodeURIComponent(project.teamSlug)}>{project.teamName}</a></li>{/each}</ul></section>
<section class="title-section"><h2>Тараулар</h2>{#each chapters.items as chapter (chapter.id)}<article class="chapter-row"><h3>{#if chapter.volume}{chapter.volume}-том / {/if}{chapter.label}-тарау{chapter.name ? ': ' + chapter.name : ''}</h3>{#each chapter.releases as release}<div class="release-row"><a class="auth-link" href={'/teams/' + encodeURIComponent(release.teamSlug)}>{release.teamName}</a><a class="button primary" href={'/read/' + encodeURIComponent(release.id)}>Оқу</a></div>{/each}</article>{:else}<p class="muted">Жарияланған тарау жоқ.</p>{/each}</section>
{#if nextURL}<a class="button subtle" href={nextURL}>Келесі тараулар</a>{/if}
<style>
 .title-header { display:flex; align-items:flex-start; gap:20px; } .title-header>div { min-width:0; } h1 { font-size:27px; line-height:1.3; overflow-wrap:anywhere; } .title-cover { width:96px; height:auto; flex-shrink:0; border-radius:8px; } dl { font-size:14px; display:grid; grid-template-columns:auto 1fr; gap:5px 15px; margin:15px 0; } dd { margin:0; } dt { color:var(--muted); } .title-genres { display:flex; flex-wrap:wrap; gap:8px; } .title-section { margin-top:30px; } h2 { font-size:21px; margin-bottom:14px; } .description { white-space:pre-wrap; overflow-wrap:anywhere; } summary { min-height:44px; padding:10px 0; cursor:pointer; } ul { padding:0; list-style:none; } .chapter-row { border-bottom:1px solid var(--line); padding:16px 0; } h3 { font-size:16px; overflow-wrap:anywhere; } .release-row { display:flex; justify-content:space-between; align-items:center; gap:16px; margin-top:8px; } @media(min-width:768px) { .title-cover { width:180px; } h1 { font-size:34px; } }
</style>
