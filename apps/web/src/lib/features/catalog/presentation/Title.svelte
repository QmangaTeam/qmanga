<script lang="ts">
 import type {Snippet} from 'svelte';
 import type {TitleDetail,Chapter,ResultPage} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {catalogNames} from './messages';
 import ChapterRow from './ChapterRow.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
 let {title,chapters,nextURL,library}:{title:TitleDetail;chapters:ResultPage<Chapter>;nextURL:string|null;library?:Snippet}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
</script>
<svelte:head><title>{title.name} — Qmanga</title><meta name="description" content={title.description.slice(0,160)}/></svelte:head>
<a class="auth-link back" href="/catalog">{text($locale,'Каталогқа оралу','В каталог')}</a>
<div class="title-header">{#if title.coverURL}<img class="title-cover" src={title.coverURL} alt="" width="280" height="400" fetchpriority="high"/>{/if}<div><p class="type">{names.kind[title.kind]}{title.demo?text($locale,' · Үлгі',' · Пример'):''}</p><h1>{title.name}</h1>{#if title.originalName}<p class="muted original">{title.originalName}</p>{/if}<dl><dt>{text($locale,'Түпнұсқа','Оригинал')}</dt><dd>{names.original[title.originalStatus]}</dd><dt>{text($locale,'Аударма','Перевод')}</dt><dd>{names.translation[title.translationStatus]}</dd></dl>{#if title.ratingCount>0}<p class="rating">{title.score.toFixed(1)} ★ <span class="muted">· {title.ratingCount} {text($locale,'баға','оценок')}</span></p>{/if}</div></div>
{#if library}<div class="library-slot">{@render library()}</div>{/if}
{#if title.description}<section class="title-section"><SectionHeading title={text($locale,'Сипаттама','Описание')}/><p class="description">{title.description}</p></section>{/if}
<div class="title-genres">{#each title.genres as genre}<ActionLink variant="ghost" href={'/catalog?genre='+encodeURIComponent(genre.slug)}>{genre.name}</ActionLink>{/each}</div>
{#if title.aliases.length}<details><summary>{text($locale,'Басқа атаулары','Другие названия')}</summary><ul>{#each title.aliases as alias}<li>{alias}</li>{/each}</ul></details>{/if}
{#if title.projects.length}<section class="title-section"><SectionHeading title={text($locale,'Аударма командалары','Команды переводчиков')}/><div class="team-links">{#each title.projects as project}<ActionLink href={'/teams/'+encodeURIComponent(project.teamSlug)}>{project.teamName}</ActionLink>{/each}</div></section>{/if}
<section class="title-section"><SectionHeading title={text($locale,'Тараулар','Главы')}/>{#each chapters.items as chapter (chapter.id)}<ChapterRow title={(chapter.volume?text($locale,chapter.volume+'-том · ','Том '+chapter.volume+' · '):'')+text($locale,chapter.label+'-тарау','Глава '+chapter.label)+(chapter.name?' · '+chapter.name:'')} releases={chapter.releases} readLabel={text($locale,'Оқу','Читать')}/>{:else}<StateMessage kind="empty" title={text($locale,'Жарияланған тарау жоқ','Опубликованных глав пока нет')}/>{/each}</section>
{#if nextURL}<Pager label={text($locale,'Тарау беттері','Страницы глав')} next={{label:text($locale,'Келесі тараулар','Следующие главы'),href:nextURL}}/>{/if}
<style>
 .back { margin:0 0 16px; } .title-header { display:flex; align-items:flex-start; gap:16px; } .title-header>div { min-width:0; } h1 { overflow-wrap:anywhere; font-size:24px; } .title-cover { width:96px; height:auto; aspect-ratio:7/10; object-fit:cover; flex-shrink:0; border-radius:6px; } .type { color:var(--muted); font-size:13px; margin-bottom:4px; } .original { margin-top:8px; } dl { font-size:14px; display:grid; grid-template-columns:auto 1fr; gap:4px 12px; margin:16px 0; } dd { margin:0; } dt { color:var(--muted); } .title-genres,.team-links { display:flex; flex-wrap:wrap; gap:8px; } .title-genres { margin-top:16px; } .title-section { margin-top:32px; } .description { white-space:pre-wrap; overflow-wrap:anywhere; max-width:760px; } summary { min-height:44px; padding:12px 0; cursor:pointer; font-size:14px; color:var(--muted); } ul { list-style:none; padding:0; font-size:14px; } .library-slot { margin-top:24px; } @media(min-width:768px) { .title-cover { width:168px; } .title-header { gap:24px; } h1 { font-size:28px; } }
</style>
