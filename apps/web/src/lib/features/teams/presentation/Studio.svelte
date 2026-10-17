<script lang="ts">
 import { onMount } from 'svelte';
 import type { TeamsGateway } from '../application/teams';
 import { requestTeam } from '../application/teams';
 import { canManage } from '../domain/team';
 import type { Team, Member } from '../domain/team';
 import type { Account } from '$lib/features/identity/domain/account';
 let { user, gateway }: { user: Account | null; gateway: TeamsGateway } = $props();
 let teams = $state<Team[]>([]);
 let selected = $state('');
 let members = $state<Member[]>([]);
 let name = $state(''); let slug = $state(''); let description = $state('');
 let email = $state(''); let role = $state<'member' | 'editor'>('member');
 let busy = $state(false); let error = $state(''); let notice = $state('');
 const labels = { owner: 'Иесі', editor: 'Редактор', member: 'Қатысушы' };
 let active = $derived(teams.find(team => team.id === selected));
 function message(failure: unknown) { error = failure instanceof Error ? failure.message : 'Қосылым үзілді'; }
 async function load() { teams = await gateway.list(); if (!selected && teams.length) selected = teams[0].id; await loadMembers(); }
 async function loadMembers() { members = []; if (active && canManage(active)) members = await gateway.members(active.id); }
 onMount(() => { if (user) load().catch(message); });
 async function create(event: SubmitEvent) {
  event.preventDefault(); busy = true; error = ''; notice = '';
  try { const team = await requestTeam(gateway, { name, slug, description }); selected = team.id; await load(); name = ''; slug = ''; description = ''; notice = 'Команда өтінімі жіберілді. Әкімші мақұлдағаннан кейін тарау жүктей аласыз.'; } catch (failure) { message(failure); } finally { busy = false; }
 }
 async function invite(event: SubmitEvent) {
  event.preventDefault(); if (!active) return; busy = true; error = '';
  try { await gateway.invite(active.id, email, role); notice = 'Шақыру жіберілді'; email = ''; } catch (failure) { message(failure); } finally { busy = false; }
 }
 async function remove(userId: string) {
  if (!active) return; busy = true;
  try { await gateway.remove(active.id, userId); await loadMembers(); } catch (failure) { message(failure); } finally { busy = false; }
 }
</script>
<svelte:head><title>Аудармашы кабинеті — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="section-heading"><div><span class="eyebrow">Бірге аударайық</span><h1>Аудармашы кабинеті</h1></div></div>
<a class="auth-link" href="/studio/projects">Аударма жобаларым</a>
{#if !user}<div class="empty-state"><h2>Алдымен есептікжазбаңа кір</h2><p>Команда құру және аударма жүктеу үшін поштаң расталуы керек.</p><a href="/login" class="button primary">Кіру</a></div>
{:else}
 {#if !user.verified}<p class="form-message">Поштаңды растау қажет. <a href="/settings">Растау хатын қайта жіберу →</a></p>{/if}
 {#if error}<p class="form-message" role="alert">{error}</p>{/if}{#if notice}<p class="form-message success" role="status">{notice}</p>{/if}
 <div class="workspace-grid">
  <section class="panel"><h2>Менің командаларым</h2>
   {#if teams.length}<label>Команда<select bind:value={selected} onchange={() => loadMembers().catch(message)}>{#each teams as team}<option value={team.id}>{team.name}</option>{/each}</select></label>
   {:else}<p class="muted">Әлі командаға қосылмадың. Жаңа командаға өтінім бере аласың.</p>{/if}
   {#if active}<div class="team-state"><span class:approved={active.approved} class="status-badge">{active.approved ? 'Мақұлданған' : 'Мақұлдау күтілуде'}</span><span>{active.role ? labels[active.role] : ''}</span></div><p class="muted">{active.description}</p>
    {#if canManage(active)}
     <h3>Қатысушылар</h3>
     {#each members as member}<div class="member-row"><div><strong>{member.email}</strong><small>{labels[member.role]}</small></div>{#if member.role !== 'owner'}<button class="button subtle" disabled={busy} onclick={() => remove(member.userId)}>Шығару</button>{/if}</div>{/each}
     <form onsubmit={invite}><h3>Шақыру жіберу</h3><label>Пошта<input type="email" bind:value={email} required autocomplete="off"/></label><label>Рөл<select bind:value={role}><option value="member">Қатысушы</option><option value="editor">Редактор</option></select></label><button class="button primary" disabled={busy}>Шақыру</button></form>
    {/if}
   {/if}
  </section>
  <section class="panel"><h2>Жаңа команда</h2><form onsubmit={create}><label>Команда атауы<input bind:value={name} minlength="2" maxlength="80" required placeholder="Дала тобы"/></label><label>Қысқа сілтеме<input bind:value={slug} pattern="[a-z0-9]+(-[a-z0-9]+)*" minlength="3" maxlength="64" required placeholder="dala-team" autocomplete="off"/></label><label>Команда туралы<textarea bind:value={description} maxlength="2000" rows="4"></textarea></label><button class="button primary" disabled={busy || !user.verified}>Өтінім жіберу</button></form></section>
 </div>
{/if}
