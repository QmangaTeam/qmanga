<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import {publicationStates} from './messages';
 const stateNames=$derived(publicationStates($locale));
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import { UploadQueue, type PublishingGateway, type Revision, type ChapterChoice } from '../application/publishing';
 import { canEdit, canSubmit } from '../domain/revision';
 let { user, projectId, gateway }: { user: Account | null; projectId: string; gateway: PublishingGateway } = $props();
 let chapters = $state<ChapterChoice[]>([]), revisions = $state<Revision[]>([]), active = $state<Revision | null>(null), chapterId = $state('');
 let busy = $state(false), uploading = $state(false), progress = $state(0), progressText = $state(''), error = $state(''), notice = $state('');
 let queue: UploadQueue | null = null;
 async function run(action: () => Promise<void>) { if (busy) return; busy = true; error = ''; notice = ''; try { await action(); } catch (failure) { error = failureCode(failure); } finally { busy = false; } }
 async function reload() { [chapters,revisions] = await Promise.all([gateway.chapters(projectId),gateway.revisions(projectId)]); }
 async function refresh() { if (active) active = await gateway.revision(active.id); }
 onMount(() => {
  if (user) run(reload);
  const timer = setInterval(() => { if (active?.state === 'processing' && !busy && !uploading && document.visibilityState === 'visible') refresh().catch(() => { error = "refresh_failed"; }); }, 2500);
  return () => { clearInterval(timer); queue?.cancel(); };
 });
 async function upload(event: Event) {
  const files = Array.from((event.currentTarget as HTMLInputElement).files || []);
  if (!active || !files.length || uploading) return;
  uploading = true; error = ''; queue = new UploadQueue(gateway);
  try { const sources = await gateway.prepare(files); await queue.run(active,sources,(name,count,total,fraction) => { progress = (count + fraction) / total; progressText = `${count}/${total}: ${name}`; }); await refresh(); notice = "upload_received"; }
  catch (failure) { error = failureCode(failure); await refresh().catch(() => {}); }
  finally { uploading = false; }
 }
 async function move(index: number, direction: number) { if (!active) return; const ids = active.pages.map(page => page.id); const other = index + direction; if (other < 0 || other >= ids.length) return; [ids[index],ids[other]] = [ids[other],ids[index]]; await gateway.reorder(active.id,active.version,ids); await refresh(); }
</script>
<svelte:head><title>{copy($locale,"Аударма жүктеу — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<a class="auth-link" href="/studio/projects">{copy($locale,"Жобаларға оралу")}</a><h1>{copy($locale,"Аударма жүктеу")}</h1>
{#if !user}<a class="button primary" href="/login">{copy($locale,"Кіру")}</a>{:else}
{#if error}<p class="form-message" role="alert">{noticeText($locale,error)}</p>{/if}{#if notice}<p class="form-message success" role="status">{noticeText($locale,notice)}</p>{/if}
<div class="workspace-grid"><section class="panel"><h2>{copy($locale,"Тарау")}</h2><form onsubmit={event => { event.preventDefault(); run(async () => { active = await gateway.create(projectId,chapterId); await reload(); }); }}><label>{copy($locale,"Тарауды таңдаңыз")}<select bind:value={chapterId} required><option value="">{copy($locale,"Тарау")}</option>{#each chapters as chapter}<option value={chapter.id}>{chapter.label} {chapter.name}</option>{/each}</select></label><button class="button primary" disabled={busy || !chapterId}>{copy($locale,"Қаралым ашу")}</button></form><p class="muted">{copy($locale,"Тарау жоқ болса, әкімшіден каталогқа қосуды сұраңыз. Жарияланған нұсқа түзету кезінде оқуға ашық қалады.")}</p>
<h3>{copy($locale,"Нұсқалар")}</h3>{#each revisions as revision}<button class="revision-option button subtle" disabled={busy || uploading} onclick={() => run(async () => { active = await gateway.revision(revision.id); })}>{revision.chapterLabel}: {stateNames[revision.state] || copy($locale,"Белгісіз")}</button>{:else}<p class="muted">{copy($locale,"Нұсқа әлі жоқ.")}</p>{/each}</section>
<section class="panel"><h2>{copy($locale,"Беттер")}</h2>{#if active}<p>{active.titleName} / {active.chapterLabel}</p><p role="status">{stateNames[active.state] || copy($locale,"Белгісіз")}</p>{#if active.note}<p class="form-message">{active.note}</p>{/if}
{#if canEdit(active.state)}<label>{copy($locale,"JPEG, PNG, WebP: 25 MiB/файл, 300 бетке дейін")}<input type="file" multiple accept="image/jpeg,image/png,image/webp" disabled={busy || uploading} onchange={upload}/></label><p class="muted">{copy($locale,"Файлдар табиғи ретпен жүктеледі. Байланыс үзілсе, сол файлдарды қайта таңдаңыз: дайындары қайталанбайды.")}</p><p class="muted">{copy($locale,"Оқу нұсқалары автоматты түрде WebP пішіміне өңделеді: үнемді, қалыпты және жоғары сапа. Кішкентай сурет үлкейтілмейді. Толық өлшемді PNG тек оқырман таңдағанда ашылады.")}</p>{/if}
{#if uploading}<progress max="1" value={progress}></progress><p class="muted">{progressText}</p><button class="button subtle" onclick={() => queue?.cancel()}>{copy($locale,"Жүктеуді тоқтату")}</button>{/if}
<ol>{#each active.pages as page,index (page.id)}<li><div><strong>{text($locale,page.position+'-бет','Страница '+page.position)}</strong><small>{page.width}×{page.height} / {(page.bytes / 1048576).toFixed(2)} MiB / {stateNames[page.status] || copy($locale,"Белгісіз")}</small></div>{#if canEdit(active.state)}<div class="row-actions"><button class="button subtle" aria-label={text($locale,page.position+'-бетті жоғары жылжыту','Переместить страницу '+page.position+' вверх')} disabled={busy || uploading || index === 0} onclick={() => run(() => move(index,-1))}>↑</button><button class="button subtle" aria-label={text($locale,page.position+'-бетті төмен жылжыту','Переместить страницу '+page.position+' вниз')} disabled={busy || uploading || index === active!.pages.length - 1} onclick={() => run(() => move(index,1))}>↓</button><button class="button subtle" disabled={busy || uploading} onclick={() => run(async () => { await gateway.remove(active!.id,page.id); await refresh(); })}>{copy($locale,"Жою")}</button></div>{/if}</li>{/each}</ol>
<div class="actions"><button class="button subtle" disabled={busy} onclick={() => run(refresh)}>{copy($locale,"Күйін жаңарту")}</button><a class="button subtle" href={'/read/' + encodeURIComponent(active.releaseId) + '?preview=' + encodeURIComponent(active.id)}>{copy($locale,"Алдын ала оқу")}</a><button class="button primary" disabled={busy || uploading || !canSubmit(active)} onclick={() => run(async () => { await gateway.submit(active!.id,active!.version); await refresh(); await reload(); notice = "submitted"; })}>{copy($locale,"Тексеруге жіберу")}</button></div>
{:else}<p class="muted">{copy($locale,"Тарауды немесе нұсқаны таңдаңыз.")}</p>{/if}</section></div>{/if}
<style>.revision-option { width:100%; justify-content:flex-start; margin:5px 0; } ol { padding:0; list-style:none; } li { border-bottom:1px solid var(--line); padding:14px 0; display:grid; gap:8px; } small { display:block; color:var(--muted); } .row-actions,.actions { display:flex; gap:8px; flex-wrap:wrap; } .actions { margin-top:20px; } progress { width:100%; } .panel label { margin:16px 0; } h1 { overflow-wrap:anywhere; }</style>
