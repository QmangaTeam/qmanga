<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale}=useLocale();
 import { page } from '$app/stores';
 import { createTeams } from '$lib/composition/browser';
 import type { PageData } from './$types';
 let { data }: { data: PageData } = $props();
 let busy = $state(false); let error = $state(''); let accepted = $state(false);
 async function accept() {
  busy = true; error = '';
  try { await createTeams().accept($page.url.searchParams.get('token') || ''); accepted = true; } catch (failure) { error = failureCode(failure); } finally { busy = false; }
 }
</script>
<svelte:head><title>{text($locale,'Командаға шақыру','Приглашение в команду')} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<div class="auth-page"><h1>{text($locale,'Командаға шақыру','Приглашение в команду')}</h1>{#if accepted}<p class="form-message success">{text($locale,'Командаға қосылдың!','Ты присоединился к команде!')}</p><a class="button primary" href="/studio">{text($locale,'Кабинетке өту','Перейти в кабинет')}</a>{:else if !data.user}<p>{text($locale,'Шақыру жіберілген поштамен кіріп, осы сілтемені қайта аш.','Войди с почтой, на которую пришло приглашение, и открой эту ссылку снова.')}</p><a class="button primary" href="/login">{text($locale,'Кіру','Войти')}</a>{:else if !data.user.verified}<p>{text($locale,'Алдымен поштаңды раста.','Сначала подтверди почту.')}</p><a class="button primary" href="/settings">Аккаунт</a>{:else}<p>{text($locale,'Шақыру тек өз поштаңа жіберілсе қабылданады.','Можно принять только приглашение, отправленное на твою почту.')}</p>{#if error}<p class="form-message" role="alert">{noticeText($locale,error)}</p>{/if}<button class="button primary" disabled={busy} onclick={accept}>{text($locale,'Шақыруды қабылдау','Принять приглашение')}</button>{/if}</div>
