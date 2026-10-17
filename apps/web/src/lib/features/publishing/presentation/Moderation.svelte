<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { PublishingGateway, Revision } from '../application/publishing';
 let { user, gateway }: { user: Account | null; gateway: PublishingGateway } = $props();
 let items = $state<Revision[]>([]), reason = $state(''), releaseId = $state(''), busy = $state(false), message = $state('');
 const allowed = $derived(user?.role === 'admin' || user?.role === 'moderator');
 async function run(action: () => Promise<void>) { if (busy) return; busy = true; message = ''; try { await action(); items = await gateway.pending(); } catch (error) { message = failureCode(error); } finally { busy = false; } }
 onMount(() => { if (allowed) run(async () => {}); });
</script>
<svelte:head><title>{copy($locale,"Жарияланымдарды тексеру — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<h1>{copy($locale,"Жарияланымдарды тексеру")}</h1>{#if user?.role === 'admin'}<a class="auth-link" href="/admin/catalog">{copy($locale,"Каталогты басқару")}</a>{/if}
{#if !allowed}<p class="form-message">{copy($locale,"Модератор рұқсаты қажет.")}</p>{:else}
{#if message}<p class="form-message" role="alert">{noticeText($locale,message)}</p>{/if}
<div class="panel"><label>{copy($locale,"Түзету немесе оқудан алу себебі")}<textarea bind:value={reason} maxlength="2000" rows="3"></textarea></label></div>
<div class="workspace-grid">{#each items as revision}<article class="panel"><h2>{revision.titleName}</h2><p>{text($locale,revision.chapterLabel+'-тарау','Глава '+revision.chapterLabel)} / {revision.teamName}</p><a class="button subtle" href={'/read/' + encodeURIComponent(revision.releaseId) + '?preview=' + encodeURIComponent(revision.id)}>{copy($locale,"Тексеріп оқу")}</a><dl><dt>{text($locale,'Дереккөз және құқық негізі','Источник и основание прав')}</dt><dd>{revision.source||'—'}</dd><dt>{text($locale,'Кредиттер','Кредиты')}</dt><dd>{revision.credits||'—'}</dd><dt>{text($locale,'Растау','Подтверждение')}</dt><dd>{revision.rightsConfirmed?text($locale,'Расталған','Подтверждено'):text($locale,'Расталмаған','Не подтверждено')}</dd></dl><div class="actions"><button class="button primary" disabled={busy} onclick={() => run(() => gateway.review(revision.id,'publish',reason))}>{copy($locale,"Жариялау")}</button><button class="button subtle" disabled={busy || !reason.trim()} onclick={() => run(() => gateway.review(revision.id,'reject',reason))}>{copy($locale,"Түзетуге қайтару")}</button></div></article>{:else}<p>{copy($locale,"Тексерілетін жаңа нұсқа жоқ.")}</p>{/each}</div>
<section class="panel"><h2>{copy($locale,"Жарияланымды оқудан алу")}</h2><form onsubmit={event => { event.preventDefault(); run(async () => { await gateway.withdraw(releaseId,reason); releaseId = ''; }); }}><label>{copy($locale,"Жарияланым ID")}<input bind:value={releaseId} required maxlength="36"/></label><button class="button subtle" disabled={busy || !reason.trim()}>{copy($locale,"Оқудан алу")}</button></form></section>{/if}
<style>dd{margin:6px 0 14px;white-space:pre-wrap;overflow-wrap:anywhere}dt{color:var(--muted);font-size:13px}.actions { display:flex; flex-wrap:wrap; gap:10px; margin-top:15px; }</style>
