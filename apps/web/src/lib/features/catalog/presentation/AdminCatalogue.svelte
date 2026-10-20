<script lang="ts">
 import {untrack} from 'svelte';
 import UnsavedChanges from '$lib/shared/presentation/ui/UnsavedChanges.svelte';
 import {fieldValidation} from '$lib/shared/presentation/form-validation';
 import {ageRatings} from '../domain/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import { editableMetadata, type CatalogManagementGateway } from '../application/management';
 import type { Genre, Summary, TitleDetail } from '../application/catalog';
 import { emptyMetadata, metadataError } from '../domain/metadata';
 import {catalogNames} from './messages';
 import PublicationFields from './PublicationFields.svelte';
 import TagSearch from './TagSearch.svelte';
 import {emptyPublication} from '../domain/publication';
 const kindNames=$derived(catalogNames($locale).kind),originalStatusNames=$derived(catalogNames($locale).original),translationStatusNames=$derived(catalogNames($locale).translation);
 let { user, gateway, initialTitleId = '' }: { user: Account | null; gateway: CatalogManagementGateway; initialTitleId?: string } = $props();
 let categories=$state<Genre[]>([]), categoryIds=$state<string[]>([]);
 let titles = $state<Summary[]>([]), genres = $state<Genre[]>([]), teams = $state<{id: string; name: string}[]>([]);
 let titleQuery=$state(''),appliedTitleQuery=$state('');
 let guard:UnsavedChanges;
 let savedMetadata=$state('');
 async function searchTitles(){if(busy)return;await operation(async()=>{const next=await gateway.titles(0,titleQuery);titles=next;moreTitles=next.length===100;appliedTitleQuery=titleQuery;});}
 let loadedAccount=untrack(()=>user?.id);
 $effect(()=>{if(user?.id!==loadedAccount){loadedAccount=user?.id;input=emptyMetadata();aliases='';authors='';categoryIds=[];selected='';detail=null;savedMetadata='';chapterLabel='';chapterName='';volume='';genreName='';genreSlug='';teamId='';titles=[];teams=[];}});
 let moreTitles=$state(false),moreTeams=$state(false);
 async function loadMoreTitles(){const next=await gateway.titles(titles.length,appliedTitleQuery);titles=[...titles,...next];moreTitles=next.length===100;}
 async function loadMoreTeams(){const next=await gateway.teams(teams.length);teams=[...teams,...next];moreTeams=next.length===100;}
 let selected = $state(''), input = $state(emptyMetadata()), aliases = $state(''), authors = $state(''), detail = $state<TitleDetail | null>(null);
 const genreSlugs=$derived(input.genreIds.map(id=>genres.find(genre=>genre.id===id)?.slug||'').filter(Boolean));
 function setGenreSlugs(slugs:string[]){input.genreIds=slugs.map(slug=>genres.find(genre=>genre.slug===slug)?.id||'').filter(Boolean);}
 let busy = $state(false), notice = $state(''), failure = $state('');
 let chapterLabel = $state(''), chapterOrder = $state(1), volume = $state(''), chapterName = $state(''), teamId = $state(''), genreName = $state(''), genreSlug = $state('');
 const metadataSnapshot=$derived(JSON.stringify({input,aliases,authors,categoryIds}));
 const dirty=$derived(savedMetadata!==''&&metadataSnapshot!==savedMetadata||chapterLabel!==''||chapterName!==''||volume!==''||genreName!==''||genreSlug!==''||teamId!=='');
 async function reload() { categories=(await gateway.taxonomy()).categories; [titles, genres, teams] = await Promise.all([gateway.titles(), gateway.genres(), gateway.teams()]);moreTitles=titles.length===100;moreTeams=teams.length===100; }
 async function operation(action: () => Promise<void>, success = '') { if (busy) return; busy = true; failure = ''; notice = ''; try { await action(); notice = success; } catch (error) { failure = failureCode(error); } finally { busy = false; } }
 async function choose(id: string) { const next = id ? await gateway.detail(id) : null; selected = id; detail = next;chapterLabel='';chapterName='';volume='';teamId='';chapterOrder=1;categoryIds=(next?.categories||[]).map(item=>item.id); input = detail ? editableMetadata(detail) : emptyMetadata(); input.publication??=emptyPublication(); aliases = input.aliases.join('\n'); authors=(input.authors||[]).join('\n'); savedMetadata=JSON.stringify({input,aliases,authors,categoryIds}); }
 onMount(() => { if (user?.role === 'admin') operation(async () => { await reload(); await choose(initialTitleId); }); });
 async function save(event: SubmitEvent) { event.preventDefault(); notice = ''; input.authors=authors.split('\n').map(value=>value.trim()).filter(Boolean); input.aliases = aliases.split('\n').map(value => value.trim()).filter(Boolean); if (metadataError(input)) { failure = "metadata_invalid"; return; } await operation(async () => { const id = await gateway.save(input, selected || undefined); await gateway.setCategories(id,categoryIds); await reload(); await choose(id); }, 'saved'); }
</script>
<UnsavedChanges bind:this={guard} {dirty}/>
<svelte:head><title>{copy($locale,"Каталогты басқару — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<h1>{copy($locale,"Каталогты басқару")}</h1><a class="auth-link" href="/admin">{copy($locale,"Командалар")}</a><a class="auth-link" href="/moderation">{copy($locale,"Жарияланымдарды тексеру")}</a>
{#if user?.role !== 'admin'}<p class="form-message">{copy($locale,"Әкімші рұқсаты қажет.")}</p>{:else}
{#if failure}<p class="form-message" role="alert">{noticeText($locale,failure)}</p>{/if}{#if notice}<p class="form-message success" role="status">{noticeText($locale,notice)}</p>{/if}
<div class="workspace-grid"><section class="panel"><h2>{copy($locale,"Тайтл")}</h2>
<form use:fieldValidation={{locale:$locale}} onsubmit={event=>{event.preventDefault();void searchTitles();}}><label>{text($locale,'Тайтл іздеу','Поиск тайтла')}<input type="search" bind:value={titleQuery} maxlength="160"/></label><button class="button subtle" disabled={busy}>{text($locale,'Іздеу','Найти')}</button></form>
<label>{copy($locale,"Таңдау")}<select value={selected} disabled={busy} onchange={event => { const picker = event.currentTarget; if(!guard.confirmDiscard()){picker.value=selected;return;} void operation(() => choose(picker.value)).finally(() => { picker.value = selected; }); }}><option value="">{copy($locale,"Жаңа тайтл")}</option>{#each titles as title}<option value={title.id}>{title.name}</option>{/each}</select></label>
{#if moreTitles}<button class="button subtle" disabled={busy} onclick={()=>operation(loadMoreTitles)}>{text($locale,'Тағы тайтлдар','Ещё тайтлы')}</button>{/if}
<form use:fieldValidation={{locale:$locale}} onsubmit={save} oninput={() => { notice = ''; }}>
 <label>{copy($locale,"Қазақша атауы")}<input bind:value={input.name} required maxlength="200"/></label>
 <label>{copy($locale,"Сілтемедегі атауы")}<input bind:value={input.slug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="128" placeholder="dala-sapary"/></label>
 <label>{copy($locale,"Түпнұсқа атауы")}<input bind:value={input.originalName} maxlength="200"/></label>
 <label>{copy($locale,"Басқа атаулар, әр жолға біреуі")}<textarea bind:value={aliases} rows="3"></textarea></label>
 <label>{text($locale,'Авторлар, әр жолға біреуі','Авторы, по одному на строке')}<textarea bind:value={authors} rows="3" maxlength="2000"></textarea></label>
 <label class="check"><input type="checkbox" bind:checked={input.listed}/>{text($locale,'Тараусыз тайтлды анонс ретінде көрсету','Показывать тайтл без глав как анонс')}</label><p class="muted">{text($locale,'Өшірілсе, тайтл алғашқы тарау жарияланғанша жасырын қалады.','Если выключено, тайтл скрыт до публикации первой главы.')}</p>
 <label class="check"><input type="checkbox" bind:checked={input.featured}/>{text($locale,'Редакция таңдауы','Подборка редакции')}</label>
 <label>{copy($locale,"Түрі")}<select bind:value={input.kind}>{#each Object.entries(kindNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{copy($locale,"Түпнұсқа күйі")}<select bind:value={input.originalStatus}>{#each Object.entries(originalStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{copy($locale,"Аударма күйі")}<select bind:value={input.translationStatus}>{#each Object.entries(translationStatusNames) as [value,name]}<option {value}>{name}</option>{/each}</select></label>
 <label>{text($locale,'Жас санаты','Возрастная маркировка')}<select name="age_rating" bind:value={input.ageRating}>{#each ageRatings as age}<option value={age}>{age===0?text($locale,'Көрсетілмеген','Не указано'):age+'+'}</option>{/each}</select></label>
 <label>{copy($locale,"Оқу бағыты")}<select bind:value={input.direction}><option value="ltr">{copy($locale,"Солдан оңға")}</option><option value="rtl">{copy($locale,"Оңнан солға")}</option></select></label>
 <label>{copy($locale,"Сипаттама")}<textarea bind:value={input.description} rows="5" maxlength="10000"></textarea></label>
 <PublicationFields bind:value={input.publication}/>
 <fieldset><legend>{copy($locale,"Жанрлар")}</legend><TagSearch {genres} {categories} field="" selected={genreSlugs} onchange={setGenreSlugs}/></fieldset>
 <fieldset><legend>{text($locale,"Мазмұн санаттары","Категории содержания")}</legend>{#each categories as category}<label class="check"><input type="checkbox" bind:group={categoryIds} value={category.id}/>{category.name}</label>{/each}</fieldset>
 <label class="check"><input type="checkbox" bind:checked={input.demo}/>{copy($locale,"Үлгі туынды")}</label>
 <button class="button primary" disabled={busy} type="submit">{copy($locale,"Тайтлды сақтау")}</button>
</form></section>
<section class="panel"><h2>{copy($locale,"Тараулар мен жобалар")}</h2>
{#if selected}<form use:fieldValidation={{locale:$locale}} onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createChapter(selected,chapterLabel,chapterOrder,volume,chapterName); chapterLabel = '';chapterName='';volume=''; }, 'saved'); }}><label>{copy($locale,"Тарау белгісі")}<input bind:value={chapterLabel} required maxlength="64" placeholder={copy($locale,"1.1 немесе Арнайы")}/></label><label>{copy($locale,"Реттік саны")}<input type="number" bind:value={chapterOrder} min="1" max="1000000" required/></label><label>{copy($locale,"Том")}<input bind:value={volume} maxlength="32"/></label><label>{copy($locale,"Тарау атауы")}<input bind:value={chapterName} maxlength="200"/></label><button class="button primary" disabled={busy}>{copy($locale,"Тарау қосу")}</button></form>
<h3>{copy($locale,"Командаға жоба беру")}</h3><form use:fieldValidation={{locale:$locale}} onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createProject(selected,teamId);teamId=''; await choose(selected); }, 'saved'); }}><label>{copy($locale,"Мақұлданған команда")}<select bind:value={teamId} required><option value="">{copy($locale,"Команданы таңдаңыз")}</option>{#each teams as team}<option value={team.id}>{team.name}</option>{/each}</select></label>{#if moreTeams}<button type="button" class="button subtle" disabled={busy} onclick={()=>operation(loadMoreTeams)}>{text($locale,'Тағы командалар','Ещё команды')}</button>{/if}<button class="button primary" disabled={busy || !teamId}>{copy($locale,"Жоба құру")}</button></form>
{#each detail?.projects || [] as project}<p><a class="auth-link" href={'/admin/projects/' + encodeURIComponent(project.id)}>{project.teamName}: {project.titleName}</a></p>{/each}
{:else}<p class="muted">{copy($locale,"Алдымен тайтлды сақтаңыз немесе таңдаңыз.")}</p>{/if}
<p><a class="auth-link" href="/admin/taxonomy">{text($locale,"Жанрлар, санаттар және жинақтар","Жанры, категории и подборки")}</a></p><h3>{copy($locale,"Жаңа жанр")}</h3><form use:fieldValidation={{locale:$locale}} onsubmit={event => { event.preventDefault(); operation(async () => { await gateway.createGenre(genreName,genreSlug); genres = await gateway.genres(); genreName = ''; genreSlug = ''; }, 'saved'); }}><label>{copy($locale,"Жанр атауы")}<input bind:value={genreName} required maxlength="80"/></label><label>{copy($locale,"Сілтемесі")}<input bind:value={genreSlug} required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="64"/></label><button class="button subtle" disabled={busy}>{copy($locale,"Жанр қосу")}</button></form>
</section></div>{/if}
<style>.auth-link { margin-right:18px; } .panel form { margin-top:18px; } .check { display:flex; align-items:center; min-height:44px; gap:10px; } .check input { width:20px; height:20px; min-height:20px; flex-shrink:0; } fieldset { border:1px solid var(--line); padding:12px; }</style>
