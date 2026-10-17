<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import { onMount } from 'svelte';
 import type { TeamsGateway } from '../application/teams';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { Team } from '../domain/team';
 let { user, gateway }: { user: Account | null; gateway: TeamsGateway } = $props();
 let teams = $state<Team[]>([]); let error = $state(''); let busy = $state(false);
 onMount(() => { if (user?.role === 'admin') gateway.pending().then(value => { teams = value; }).catch(failure => { error = failureCode(failure); }); });
 async function approve(id: string) {
  busy = true; error = '';
  try { await gateway.approve(id); teams = await gateway.pending(); } catch (failure) { error = failureCode(failure); } finally { busy = false; }
 }
</script>
<svelte:head><title>{copy($locale,"Командаларды мақұлдау — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<div class="section-heading"><div><h1>{copy($locale,"Команда өтінімдері")}</h1></div></div>
<a class="auth-link" href="/admin/catalog">{copy($locale,"Каталогты басқару")}</a><a class="auth-link" href="/moderation">{copy($locale,"Жарияланымдарды тексеру")}</a>
{#if !user || user.role !== 'admin'}<div class="empty-state"><h2>{copy($locale,"Әкімші рұқсаты қажет")}</h2><a class="button primary" href="/login">{copy($locale,"Кіру")}</a></div>{:else}
 {#if error}<p class="form-message" role="alert">{noticeText($locale,error)}</p>{/if}
 <div class="workspace-grid">{#each teams as team}<article class="panel"><span class="eyebrow">{team.slug}</span><h2>{team.name}</h2><p class="muted">{team.description}</p><button class="button primary" disabled={busy} onclick={() => approve(team.id)}>{copy($locale,"Команданы мақұлдау")}</button></article>{:else}<div class="empty-state"><h2>{copy($locale,"Жаңа өтінім жоқ")}</h2><p>{copy($locale,"Командалар жіберген өтінімдер осында көрінеді.")}</p></div>{/each}</div>
{/if}
