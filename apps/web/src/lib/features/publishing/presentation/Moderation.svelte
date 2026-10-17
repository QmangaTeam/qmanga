<script lang="ts">
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { PublishingGateway, Revision } from '../application/publishing';
 let { user, gateway }: { user: Account | null; gateway: PublishingGateway } = $props();
 let items = $state<Revision[]>([]), reason = $state(''), releaseId = $state(''), busy = $state(false), message = $state('');
 const allowed = $derived(user?.role === 'admin' || user?.role === 'moderator');
 async function run(action: () => Promise<void>) { if (busy) return; busy = true; message = ''; try { await action(); items = await gateway.pending(); } catch (error) { message = error instanceof Error ? error.message : 'Қосылым үзілді'; } finally { busy = false; } }
 onMount(() => { if (allowed) run(async () => {}); });
</script>
<svelte:head><title>Жарияланымдарды тексеру — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<h1>Жарияланымдарды тексеру</h1>{#if user?.role === 'admin'}<a class="auth-link" href="/admin/catalog">Каталогты басқару</a>{/if}
{#if !allowed}<p class="form-message">Модератор рұқсаты қажет.</p>{:else}
{#if message}<p class="form-message" role="alert">{message}</p>{/if}
<div class="panel"><label>Түзету немесе оқудан алу себебі<textarea bind:value={reason} maxlength="2000" rows="3"></textarea></label></div>
<div class="workspace-grid">{#each items as revision}<article class="panel"><h2>{revision.titleName}</h2><p>{revision.chapterLabel}-тарау / {revision.teamName}</p><a class="button subtle" href={'/read/' + encodeURIComponent(revision.releaseId) + '?preview=' + encodeURIComponent(revision.id)}>Тексеріп оқу</a><div class="actions"><button class="button primary" disabled={busy} onclick={() => run(() => gateway.review(revision.id,'publish',reason))}>Жариялау</button><button class="button subtle" disabled={busy || !reason.trim()} onclick={() => run(() => gateway.review(revision.id,'reject',reason))}>Түзетуге қайтару</button></div></article>{:else}<p>Тексерілетін жаңа нұсқа жоқ.</p>{/each}</div>
<section class="panel"><h2>Жарияланымды оқудан алу</h2><form onsubmit={event => { event.preventDefault(); run(async () => { await gateway.withdraw(releaseId,reason); releaseId = ''; }); }}><label>Жарияланым ID<input bind:value={releaseId} required maxlength="36"/></label><button class="button subtle" disabled={busy || !reason.trim()}>Оқудан алу</button></form></section>{/if}
<style>.actions { display:flex; flex-wrap:wrap; gap:10px; margin-top:15px; }</style>
