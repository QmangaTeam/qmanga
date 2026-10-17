<script lang="ts">
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import { editableMetadata, type CatalogManagementGateway } from '../application/management';
 import type { Genre, Summary, TitleDetail } from '../application/catalog';
 import { emptyMetadata, metadataError } from '../domain/metadata';
 import { kindNames, originalStatusNames, translationStatusNames } from '../domain/catalog';
 let { user, gateway }: { user: Account | null; gateway: CatalogManagementGateway } = $props();
 let titles = $state<Summary[]>([]), genres = $state<Genre[]>([]), teams = $state<{id: string; name: string}[]>([]);
 let selected = $state(''), input = $state(emptyMetadata()), aliases = $state(''), detail = $state<TitleDetail | null>(null);
 let busy = $state(false), notice = $state(''), failure = $state('');
 let chapterLabel = $state(''), chapterOrder = $state(1), volume = $state(''), chapterName = $state(''), teamId = $state(''), genreName = $state(''), genreSlug = $state('');
 async function reload() { [titles, genres, teams] = await Promise.all([gateway.titles(), gateway.genres(), gateway.teams()]); }
 async function operation(action: () => Promise<void>) { if (busy) return; busy = true; failure = ''; notice = ''; try { await action(); notice = 'Сақталды'; } catch (error) { failure = error instanceof Error ? error.message : 'Қосылым үзілді'; } finally { busy = false; } }
 async function choose(id: string) { selected = id; detail = id ? await gateway.detail(id) : null; input = detail ? editableMetadata(detail) : emptyMetadata(); aliases = input.aliases.join('\n'); }
 onMount(() => { if (user?.role === 'admin') operation(reload); });
 async function save(event: SubmitEvent) { event.preventDefault(); input.aliases = aliases.split('\n').map(value => value.trim()).filter(Boolean); if (metadataError(input)) { failure = 'Атауы, сілтемесі мен мәтін ұзындығын тексеріңіз'; return; } await operation(async () => { const id = await gateway.save(input, selected || undefined); await reload(); await choose(id); }); }
</script>
<svelte:head><title>Каталогты басқару — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<h1>Каталогты басқару</h1><a class="auth-link" href="/admin">Командалар</a><a class="auth-link" href="/moderation">Жарияланымдарды тексеру</a>
{#if user?.role !== 'admin'}<p class="form-message">Әкімші рұқсаты қажет.</p>{:else}
{#if failure}<p class="form-message" role="alert">{failure}</p>{/if}{#if notice}<p class="form-message success" role="status">{notice}</p>{/if}
<div class="workspace-grid"><section class="panel"><h2>Тайтл</h2>
<label>Таңдау<select value={selected} disabled={busy} onchange={event => operation(() => choose(event.currentTarget.value))}><option value="">Жаңа тайтл</option>{#each titles as title}<option value={title.id}>{title.name}</option>{/each}</select></label>
<form onsubmit={save}>
 <label>Қазақша атауы<input bind:value={input.name} required maxlength="200"/></label>
 <label>Сілтемедегі атауы<input bind:value={input.slug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="128" placeholder="dala-sapary"/></label>
 <label>Түпнұсқа атауы<input bind:value={input.originalName} maxlength="200"/></label>
 <label>Басқа атаулар, әр жолға біреуі<textarea bind:value={aliases} rows="3"></textarea></label>
 <label>Түрі<select bind:value={input.kind}>{#each Object.entries(kindNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>Түпнұсқа күйі<select bind:value={input.originalStatus}>{#each Object.entries(originalStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>Аударма күйі<select bind:value={input.translationStatus}>{#each Object.entries(translationStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>Оқу бағыты<select bind:value={input.direction}><option value="ltr">Солдан оңға</option><option value="rtl">Оңнан солға</option></select></label>
 <label>Сипаттама<textarea bind:value={input.description} rows="5" maxlength="10000"></textarea></label>
 <fieldset><legend>Жанрлар</legend>{#each genres as genre}<label class="check"><input type="checkbox" bind:group={input.genreIds} value={genre.id}/>{genre.name}</label>{/each}</fieldset>
 <label class="check"><input type="checkbox" bind:checked={input.demo}/>Үлгі туынды</label>
 <button class="button primary" disabled={busy} type="submit">Тайтлды сақтау</button>
</form></section>
<section class="panel"><h2>Тараулар мен жобалар</h2>
{#if selected}<form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createChapter(selected,chapterLabel,chapterOrder,volume,chapterName); chapterLabel = ''; }); }}><label>Тарау белгісі<input bind:value={chapterLabel} required maxlength="64" placeholder="1.1 немесе Арнайы"/></label><label>Реттік саны<input type="number" bind:value={chapterOrder} min="1" max="1000000" required/></label><label>Том<input bind:value={volume} maxlength="32"/></label><label>Тарау атауы<input bind:value={chapterName} maxlength="200"/></label><button class="button primary" disabled={busy}>Тарау қосу</button></form>
<h3>Командаға жоба беру</h3><form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createProject(selected,teamId); await choose(selected); }); }}><label>Мақұлданған команда<select bind:value={teamId} required><option value="">Команданы таңдаңыз</option>{#each teams as team}<option value={team.id}>{team.name}</option>{/each}</select></label><button class="button primary" disabled={busy || !teamId}>Жоба құру</button></form>
{#each detail?.projects || [] as project}<p><a class="auth-link" href={'/studio/projects/' + encodeURIComponent(project.id)}>{project.teamName}: {project.titleName}</a></p>{/each}
{:else}<p class="muted">Алдымен тайтлды сақтаңыз немесе таңдаңыз.</p>{/if}
<h3>Жаңа жанр</h3><form onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createGenre(genreName,genreSlug); genres = await gateway.genres(); genreName = ''; genreSlug = ''; }); }}><label>Жанр атауы<input bind:value={genreName} required maxlength="80"/></label><label>Сілтемесі<input bind:value={genreSlug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="64"/></label><button class="button subtle" disabled={busy}>Жанр қосу</button></form>
</section></div>{/if}
<style>.auth-link { margin-right:18px; } .panel form { margin-top:18px; } .check { display:flex; align-items:center; min-height:44px; gap:10px; } .check input { width:20px; height:20px; min-height:20px; flex-shrink:0; } fieldset { border:1px solid var(--line); padding:12px; }</style>
