<script lang="ts">
 import { onMount } from 'svelte';
 import type { TeamsGateway } from '../application/teams';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { Team } from '../domain/team';
 let { user, gateway }: { user: Account | null; gateway: TeamsGateway } = $props();
 let teams = $state<Team[]>([]); let error = $state(''); let busy = $state(false);
 onMount(() => { if (user?.role === 'admin') gateway.pending().then(value => { teams = value; }).catch(failure => { error = failure instanceof Error ? failure.message : 'Қосылым үзілді'; }); });
 async function approve(id: string) {
  busy = true; error = '';
  try { await gateway.approve(id); teams = await gateway.pending(); } catch (failure) { error = failure instanceof Error ? failure.message : 'Қосылым үзілді'; } finally { busy = false; }
 }
</script>
<svelte:head><title>Командаларды мақұлдау — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="section-heading"><div><span class="eyebrow">Басқару</span><h1>Команда өтінімдері</h1></div></div>
{#if !user || user.role !== 'admin'}<div class="empty-state"><h2>Әкімші рұқсаты қажет</h2><a class="button primary" href="/login">Кіру</a></div>{:else}
 {#if error}<p class="form-message" role="alert">{error}</p>{/if}
 <div class="workspace-grid">{#each teams as team}<article class="panel"><span class="eyebrow">{team.slug}</span><h2>{team.name}</h2><p class="muted">{team.description}</p><button class="button primary" disabled={busy} onclick={() => approve(team.id)}>Команданы мақұлдау</button></article>{:else}<div class="empty-state"><h2>Жаңа өтінім жоқ</h2><p>Командалар жіберген өтінімдер осында көрінеді.</p></div>{/each}</div>
{/if}
