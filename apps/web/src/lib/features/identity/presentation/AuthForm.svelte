<script lang="ts">
 import { page } from '$app/stores';
 import { goto, invalidateAll } from '$app/navigation';
 import { authenticate } from '../application/auth';
 import type { AuthResult, AuthGateway } from '../application/auth';
 let { mode, gateway }: { mode: 'login' | 'register' | 'forgot-password' | 'reset-password' | 'verify-email'; gateway: AuthGateway } = $props();
 let email = $state('');
 let password = $state('');
 let busy = $state(false);
 let result = $state<AuthResult | null>(null);
 const titles = { login: 'Қайта оралғаныңа қуаныштымыз', register: 'Өз оқиғаңды баста', 'forgot-password': 'Құпиясөзді ұмыттың ба?', 'reset-password': 'Жаңа құпиясөз', 'verify-email': 'Поштаңды раста' };
 const labels = { login: 'Кіру', register: 'Тіркелу', 'forgot-password': 'Нұсқаулық жіберу', 'reset-password': 'Құпиясөзді жаңарту', 'verify-email': 'Поштаны растау' };
 async function submit(event: SubmitEvent) {
  event.preventDefault();
  if (busy) return;
  busy = true; result = null;
  const token = $page.url.searchParams.get('token') || '';
  try {
   if (mode === 'login' || mode === 'register') result = await authenticate(gateway, mode, { email, password });
   else if (mode === 'forgot-password') result = await gateway.requestReset(email);
   else if (mode === 'verify-email') result = await gateway.verify(token);
   else result = await gateway.reset(token, password);
   if (mode === 'login' && result.ok) { await invalidateAll(); await goto('/'); }
  } finally { busy = false; }
 }
 async function resend() {
  busy = true;
  try { result = await gateway.resend(email); } finally { busy = false; }
 }
</script>
<svelte:head><title>{labels[mode]} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="auth-page">
 <a class="eyebrow" href="/">← Qmanga</a>
 <h1>{titles[mode]}</h1>
 <p class="auth-intro">{mode === 'login' ? 'Кітапханаң мен оқыған жерің сені күтіп тұр.' : mode === 'register' ? 'Сақта, талқыла және оқуды жалғастыр.' : 'Есептікжазбаңа қолжетімділікті қауіпсіз қалпына келтір.'}</p>
 <form onsubmit={submit}>
  {#if mode === 'login' || mode === 'register' || mode === 'forgot-password'}
   <label>Электрондық пошта<input type="email" bind:value={email} autocomplete="email" inputmode="email" required maxlength="254" placeholder="sen@example.com"/></label>
  {/if}
  {#if mode === 'login' || mode === 'register' || mode === 'reset-password'}
   <label>Құпиясөз<input type="password" bind:value={password} autocomplete={mode === 'login' ? 'current-password' : 'new-password'} required minlength={mode === 'login' ? 1 : 10} maxlength="128" placeholder="Кемінде 10 таңба"/></label>
  {/if}
  {#if mode === 'verify-email'}<p>Хаттағы сілтемені ашқан болсаң, растау батырмасын бас.</p>{/if}
  {#if result}<p class:success={result.ok} class="form-message" role={result.ok ? 'status' : 'alert'}>{result.message}</p>{/if}
  <button class="button primary" type="submit" disabled={busy}>{busy ? 'Күте тұр…' : labels[mode]} <span aria-hidden="true">↗</span></button>
 </form>
 {#if mode === 'login'}<a class="auth-link" href="/forgot-password">Құпиясөзді ұмыттың ба?</a><p class="auth-bottom">Әлі тіркелмедің бе? <a href="/register">Тіркелу</a></p>
 {:else if mode === 'register'}<p class="auth-bottom">Есептікжазбаң бар ма? <a href="/login">Кіру</a></p><button class="auth-link" disabled={busy || !email} onclick={resend}>Растау хатын қайта жіберу</button>
 {:else}<a class="auth-link" href="/login">Кіру бетіне оралу</a>{/if}
 <p class="auth-note">Оқу үшін есептікжазба қажет емес. <a href="/catalog">Каталогты ашу ↗</a></p>
</div>
