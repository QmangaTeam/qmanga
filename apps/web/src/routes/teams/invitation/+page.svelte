<script lang="ts">
 import { page } from '$app/stores';
 import { createTeams } from '$lib/composition/browser';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
 let busy = $state(false); let error = $state(''); let accepted = $state(false);
 async function accept() {
  busy = true; error = '';
  try { await createTeams().accept($page.url.searchParams.get('token') || ''); accepted = true; } catch (failure) { error = failure instanceof Error ? failure.message : 'Қосылым үзілді'; } finally { busy = false; }
 }
</script>
<svelte:head><title>Командаға шақыру — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="auth-page"><span class="eyebrow">Бірге аударайық</span><h1>Командаға шақыру</h1>{#if accepted}<p class="form-message success">Командаға қосылдың!</p><a class="button primary" href="/studio">Кабинетке өту</a>{:else if !data.user}<p>Шақыру жіберілген поштамен кіріп, осы сілтемені қайта аш.</p><a class="button primary" href="/login">Кіру</a>{:else if !data.user.verified}<p>Алдымен поштаңды раста.</p><a class="button primary" href="/settings">Есептікжазба</a>{:else}<p>Шақыру тек өз поштаңа жіберілген болса қабылданады.</p>{#if error}<p class="form-message" role="alert">{error}</p>{/if}<button class="button primary" disabled={busy} onclick={accept}>Шақыруды қабылдау</button>{/if}</div>
