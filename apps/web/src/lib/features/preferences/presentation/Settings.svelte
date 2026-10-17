<script lang="ts">
 import { onMount } from 'svelte';
 import { goto, invalidateAll } from '$app/navigation';
 import type { AuthGateway } from '$lib/features/identity/application/auth';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { ThemeStore } from '../application/theme';
 import { initializeTheme, changeTheme } from '$lib/features/preferences/application/theme';
 import type { Theme } from '$lib/features/preferences/domain/theme';

 let { user, auth, preferences }: { user: Account | null; auth: AuthGateway; preferences: ThemeStore } = $props();
 let theme = $state<Theme>('system'); let busy = $state(false); let message = $state('');

 onMount(() => { theme = initializeTheme(preferences); });
 async function logout() {
  busy = true;
  try { const result = await auth.logout(); message = result.message; if (result.ok) { await invalidateAll(); await goto('/'); } } finally { busy = false; }
 }
 async function resend() { if (!user) return; busy = true; try { message = (await auth.resend(user.email)).message; } finally { busy = false; } }
</script>
<svelte:head><title>Баптаулар — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="section-heading"><div><span class="eyebrow">Өзіңе ыңғайла</span><h1>Баптаулар</h1></div></div>
<div class="workspace-grid">
 <section class="panel"><h2>Есептікжазба</h2>
  {#if user}<p><strong>{user.email}</strong></p><p class="muted">{user.verified ? 'Пошта расталған' : 'Пошта әлі расталмаған'}</p>{#if !user.verified}<button class="button subtle" disabled={busy} onclick={resend}>Растау хатын жіберу</button>{/if}
   <div class="settings-actions"><a class="button subtle" href="/studio">Аудармашы кабинеті ↗</a>{#if user.role === 'admin'}<a class="button subtle" href="/admin">Әкімші панелі ↗</a>{/if}<button class="button subtle" disabled={busy} onclick={logout}>Есептікжазбадан шығу</button></div>
  {:else}<p class="muted">Кітапхана мен оқу орнын құрылғылар арасында сақтау үшін кір.</p><a class="button primary" href="/login">Кіру</a><a class="button subtle" href="/register">Тіркелу</a>{/if}
  {#if message}<p class="form-message" role="status">{message}</p>{/if}
 </section>
 <section class="panel"><h2>Сыртқы көрініс</h2><label>Тақырып<select value={theme} onchange={event => { theme = changeTheme(preferences, event.currentTarget.value); }}><option value="system">Құрылғы тақырыбы</option><option value="dark">Қараңғы</option><option value="light">Жарық</option></select></label><p class="muted">Таңдау осы құрылғыда сақталады.</p></section>
 <section class="panel"><h2>Оқу</h2><a class="button subtle" href="/saved">Сақталған тараулар</a>{#if user?.role==='moderator'||user?.role==='admin'}<a class="button subtle" href="/moderation">Модерация</a>{/if}</section>
</div>
