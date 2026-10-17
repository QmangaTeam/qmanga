<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import { editableMetadata, type CatalogManagementGateway } from '../application/management';
 import type { Genre, Summary, TitleDetail } from '../application/catalog';
 import { emptyMetadata, metadataError } from '../domain/metadata';
 import {catalogNames} from './messages';
 const kindNames=$derived(catalogNames($locale).kind),originalStatusNames=$derived(catalogNames($locale).original),translationStatusNames=$derived(catalogNames($locale).translation);
 let { user, gateway }: { user: Account | null; gateway: CatalogManagementGateway } = $props();
 let titles = $state<Summary[]>([]), genres = $state<Genre[]>([]), teams = $state<{id: string; name: string}[]>([]);
 let moreTitles=$state(false),moreTeams=$state(false);
 async function loadMoreTitles(){const next=await gateway.titles(titles.length);titles=[...titles,...next];moreTitles=next.length===100;}
 async function loadMoreTeams(){const next=await gateway.teams(teams.length);teams=[...teams,...next];moreTeams=next.length===100;}
 let selected = $state(''), input = $state(emptyMetadata()), aliases = $state(''), authors = $state(''), detail = $state<TitleDetail | null>(null);
 let busy = $state(false), notice = $state(''), failure = $state('');
 let chapterLabel = $state(''), chapterOrder = $state(1), volume = $state(''), chapterName = $state(''), teamId = $state(''), genreName = $state(''), genreSlug = $state('');
 async function reload() { [titles, genres, teams] = await Promise.all([gateway.titles(), gateway.genres(), gateway.teams()]);moreTitles=titles.length===100;moreTeams=teams.length===100; }
 async function operation(action: () => Promise<void>) { if (busy) return; busy = true; failure = ''; notice = ''; try { await action(); notice = "saved"; } catch (error) { failure = failureCode(error); } finally { busy = false; } }
 async function choose(id: string) { selected = id; detail = id ? await gateway.detail(id) : null; input = detail ? editableMetadata(detail) : emptyMetadata(); aliases = input.aliases.join('\n'); authors=(input.authors||[]).join('\n'); }
 onMount(() => { if (user?.role === 'admin') operation(reload); });
 async function save(event: SubmitEvent) { event.preventDefault(); input.authors=authors.split('\n').map(value=>value.trim()).filter(Boolean); input.aliases = aliases.split('\n').map(value => value.trim()).filter(Boolean); if (metadataError(input)) { failure = "metadata_invalid"; return; } await operation(async () => { const id = await gateway.save(input, selected || undefined); await reload(); await choose(id); }); }
</script>
<svelte:head><title>{copy($locale,"Каталогты басқару — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<h1>{copy($locale,"Каталогты басқару")}</h1><a class="auth-link" href="/admin">{copy($locale,"Командалар")}</a><a class="auth-link" href="/moderation">{copy($locale,"Жарияланымдарды тексеру")}</a>
{#if user?.role !== 'admin'}<p class="form-message">{copy($locale,"Әкімші рұқсаты қажет.")}</p>{:else}
{#if failure}<p class="form-message" role="alert">{noticeText($locale,failure)}</p>{/if}{#if notice}<p class="form-message success" role="status">{noticeText($locale,notice)}</p>{/if}
<div class="workspace-grid"><section class="panel"><h2>{copy($locale,"Тайтл")}</h2>
<label>{copy($locale,"Таңдау")}<select value={selected} disabled={busy} onchange={event => operation(() => choose(event.currentTarget.value))}><option value="">{copy($locale,"Жаңа тайтл")}</option>{#each titles as title}<option value={title.id}>{title.name}</option>{/each}</select></label>
{#if moreTitles}<button class="button subtle" disabled={busy} onclick={()=>operation(loadMoreTitles)}>{text($locale,'Тағы тайтлдар','Ещё тайтлы')}</button>{/if}
<form onsubmit={save}>
 <label>{copy($locale,"Қазақша атауы")}<input bind:value={input.name} required maxlength="200"/></label>
 <label>{copy($locale,"Сілтемедегі атауы")}<input bind:value={input.slug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="128" placeholder="dala-sapary"/></label>
 <label>{copy($locale,"Түпнұсқа атауы")}<input bind:value={input.originalName} maxlength="200"/></label>
 <label>{copy($locale,"Басқа атаулар, әр жолға біреуі")}<textarea bind:value={aliases} rows="3"></textarea></label>
 <label>{text($locale,'Авторлар, әр жолға біреуі','Авторы, по одному на строке')}<textarea bind:value={authors} rows="3" maxlength="2000"></textarea></label>
 <label class="check"><input type="checkbox" bind:checked={input.featured}/>{text($locale,'Редакция таңдауы','Подборка редакции')}</label>
 <label>{copy($locale,"Түрі")}<select bind:value={input.kind}>{#each Object.entries(kindNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{copy($locale,"Түпнұсқа күйі")}<select bind:value={input.originalStatus}>{#each Object.entries(originalStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{copy($locale,"Аударма күйі")}<select bind:value={input.translationStatus}>{#each Object.entries(translationStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{copy($locale,"Оқу бағыты")}<select bind:value={input.direction}><option value="ltr">{copy($locale,"Солдан оңға")}</option><option value="rtl">{copy($locale,"Оңнан солға")}</option></select></label>
 <label>{copy($locale,"Сипаттама")}<textarea bind:value={input.description} rows="5" maxlength="10000"></textarea></label>
 <fieldset><legend>{copy($locale,"Жанрлар")}</legend>{#each genres as genre}<label class="check"><input type="checkbox" bind:group={input.genreIds} value={genre.id}/>{genre.name}</label>{/each}</fieldset>
 <label class="check"><input type="checkbox" bind:checked={input.demo}/>{copy($locale,"Үлгі туынды")}</label>
 <button class="button primary" disabled={busy} type="submit">{copy($locale,"Тайтлды сақтау")}</button>
</form></section>
<section class="panel"><h2>{copy($locale,"Тараулар мен жобалар")}</h2>
{#if selected}<form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createChapter(selected,chapterLabel,chapterOrder,volume,chapterName); chapterLabel = ''; }); }}><label>{copy($locale,"Тарау белгісі")}<input bind:value={chapterLabel} required maxlength="64" placeholder={copy($locale,"1.1 немесе Арнайы")}/></label><label>{copy($locale,"Реттік саны")}<input type="number" bind:value={chapterOrder} min="1" max="1000000" required/></label><label>{copy($locale,"Том")}<input bind:value={volume} maxlength="32"/></label><label>{copy($locale,"Тарау атауы")}<input bind:value={chapterName} maxlength="200"/></label><button class="button primary" disabled={busy}>{copy($locale,"Тарау қосу")}</button></form>
<h3>{copy($locale,"Командаға жоба беру")}</h3><form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createProject(selected,teamId); await choose(selected); }); }}><label>{copy($locale,"Мақұлданған команда")}<select bind:value={teamId} required><option value="">{copy($locale,"Команданы таңдаңыз")}</option>{#each teams as team}<option value={team.id}>{team.name}</option>{/each}</select></label>{#if moreTeams}<button type="button" class="button subtle" disabled={busy} onclick={()=>operation(loadMoreTeams)}>{text($locale,'Тағы командалар','Ещё команды')}</button>{/if}<button class="button primary" disabled={busy || !teamId}>{copy($locale,"Жоба құру")}</button></form>
{#each detail?.projects || [] as project}<p><a class="auth-link" href={'/studio/projects/' + encodeURIComponent(project.id)}>{project.teamName}: {project.titleName}</a></p>{/each}
{:else}<p class="muted">{copy($locale,"Алдымен тайтлды сақтаңыз немесе таңдаңыз.")}</p>{/if}
<h3>{copy($locale,"Жаңа жанр")}</h3><form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createGenre(genreName,genreSlug); genres = await gateway.genres(); genreName = ''; genreSlug = ''; }); }}><label>{copy($locale,"Жанр атауы")}<input bind:value={genreName} required maxlength="80"/></label><label>{copy($locale,"Сілтемесі")}<input bind:value={genreSlug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="64"/></label><button class="button subtle" disabled={busy}>{copy($locale,"Жанр қосу")}</button></form>
</section></div>{/if}
<style>.auth-link { margin-right:18px; } .panel form { margin-top:18px; } .check { display:flex; align-items:center; min-height:44px; gap:10px; } .check input { width:20px; height:20px; min-height:20px; flex-shrink:0; } fieldset { border:1px solid var(--line); padding:12px; }</style>
