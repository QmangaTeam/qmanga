<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
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
 const labels = $derived({ owner: copy($locale,"Иесі"), editor: copy($locale,"Редактор"), member: copy($locale,"Қатысушы") });
 let active = $derived(teams.find(team => team.id === selected));
 function message(failure: unknown) { error = failureCode(failure); }
 async function load() { teams = await gateway.list(); if (!selected && teams.length) selected = teams[0].id; await loadMembers(); }
 async function loadMembers() { members = []; if (active && canManage(active)) members = await gateway.members(active.id); }
 onMount(() => { if (user) load().catch(message); });
 async function create(event: SubmitEvent) {
  event.preventDefault(); busy = true; error = ''; notice = '';
  try { const team = await requestTeam(gateway, { name, slug, description }); selected = team.id; await load(); name = ''; slug = ''; description = ''; notice = "team_requested"; } catch (failure) { message(failure); } finally { busy = false; }
 }
 async function invite(event: SubmitEvent) {
  event.preventDefault(); if (!active) return; busy = true; error = '';
  try { await gateway.invite(active.id, email, role); notice = "invite_sent"; email = ''; } catch (failure) { message(failure); } finally { busy = false; }
 }
 async function remove(userId: string) {
  if (!active) return; busy = true;
  try { await gateway.remove(active.id, userId); await loadMembers(); } catch (failure) { message(failure); } finally { busy = false; }
 }
</script>
<svelte:head><title>{copy($locale,"Аудармашы кабинеті — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<div class="section-heading"><div><h1>{copy($locale,"Аудармашы кабинеті")}</h1></div></div>
<a class="auth-link" href="/studio/projects">{copy($locale,"Аударма жобаларым")}</a>
{#if !user}<div class="empty-state"><h2>{copy($locale,"Алдымен есептікжазбаңа кір")}</h2><p>{copy($locale,"Команда құру және аударма жүктеу үшін поштаң расталуы керек.")}</p><a href="/login" class="button primary">{copy($locale,"Кіру")}</a></div>
{:else}
 {#if !user.verified}<p class="form-message">{copy($locale,"Поштаңды растау қажет. ")}<a href="/settings">{copy($locale,"Растау хатын қайта жіберу →")}</a></p>{/if}
 {#if error}<p class="form-message" role="alert">{noticeText($locale,error)}</p>{/if}{#if notice}<p class="form-message success" role="status">{noticeText($locale,notice)}</p>{/if}
 <div class="workspace-grid">
  <section class="panel"><h2>{copy($locale,"Менің командаларым")}</h2>
   {#if teams.length}<label>{copy($locale,"Команда")}<select bind:value={selected} onchange={() => loadMembers().catch(message)}>{#each teams as team}<option value={team.id}>{team.name}</option>{/each}</select></label>
   {:else}<p class="muted">{copy($locale,"Әлі командаға қосылмадың. Жаңа командаға өтінім бере аласың.")}</p>{/if}
   {#if active}<div class="team-state"><span class:approved={active.approved} class="status-badge">{active.approved ? copy($locale,"Мақұлданған") : copy($locale,"Мақұлдау күтілуде")}</span><span>{active.role ? labels[active.role] : ''}</span></div><p class="muted">{active.description}</p>
    {#if canManage(active)}
     <h3>{copy($locale,"Қатысушылар")}</h3>
     {#each members as member}<div class="member-row"><div><strong>{member.email}</strong><small>{labels[member.role]}</small></div>{#if member.role !== 'owner'}<button class="button subtle" disabled={busy} onclick={() => remove(member.userId)}>{copy($locale,"Шығару")}</button>{/if}</div>{/each}
     <form onsubmit={invite}><h3>{copy($locale,"Шақыру жіберу")}</h3><label>{copy($locale,"Пошта")}<input type="email" bind:value={email} required autocomplete="off"/></label><label>{copy($locale,"Рөл")}<select bind:value={role}><option value="member">{copy($locale,"Қатысушы")}</option><option value="editor">{copy($locale,"Редактор")}</option></select></label><button class="button primary" disabled={busy}>{copy($locale,"Шақыру")}</button></form>
    {/if}
   {/if}
  </section>
  <section class="panel"><h2>{copy($locale,"Жаңа команда")}</h2><form onsubmit={create}><label>{copy($locale,"Команда атауы")}<input bind:value={name} minlength="2" maxlength="80" required placeholder={copy($locale,"Дала тобы")}/></label><label>{copy($locale,"Қысқа сілтеме")}<input bind:value={slug} pattern="[a-z0-9]+(-[a-z0-9]+)*" minlength="3" maxlength="64" required placeholder="dala-team" autocomplete="off"/></label><label>{copy($locale,"Команда туралы")}<textarea bind:value={description} maxlength="2000" rows="4"></textarea></label><button class="button primary" disabled={busy || !user.verified}>{copy($locale,"Өтінім жіберу")}</button></form></section>
 </div>
{/if}
