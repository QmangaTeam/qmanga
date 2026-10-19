<script lang="ts">
 import {onMount} from 'svelte';
 import {page} from '$app/stores';
 import {goto,invalidateAll} from '$app/navigation';
 import {authenticate,type AuthResult,type AuthGateway} from '../application/auth';
 import {safeReturnPath} from '../domain/return-path';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {errorText} from '$lib/shared/presentation/messages';
 import AuthLayout from '$lib/shared/presentation/ui/AuthLayout.svelte';
 import Field from '$lib/shared/presentation/ui/Field.svelte';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {mode,gateway}:{mode:'login'|'register'|'forgot-password'|'reset-password'|'verify-email';gateway:AuthGateway}=$props();
 const {locale}=useLocale();
 let googleEnabled=$state(false);
 onMount(async()=>{googleEnabled=await gateway.googleStatus?.()||false;});
 async function googleLogin(){if(busy||!gateway.googleStart)return;busy=true;try{const response=await gateway.googleStart(safeReturnPath($page.url.searchParams.get('returnTo')),'');if(response.url)globalThis.location.assign(response.url);else result={ok:false,code:response.code,message:''};}finally{busy=false;}}
 let displayName=$state(''),email=$state(''),password=$state(''),busy=$state(false),result=$state<AuthResult|null>(null);
 const labels=$derived({login:text($locale,'Кіру','Войти'),register:text($locale,'Тіркелу','Регистрация'),'forgot-password':text($locale,'Құпиясөзді қалпына келтіру','Восстановить пароль'),'reset-password':text($locale,'Жаңа құпиясөз','Новый пароль'),'verify-email':text($locale,'Поштаны растау','Подтвердить почту')});
 const success=$derived(mode==='register'||mode==='forgot-password'?text($locale,'Хат жіберілді. Поштаңды тексер.','Письмо отправлено. Проверь почту.'):text($locale,'Дайын. Аккаунтқа кіруге болады.','Готово. Можно войти в аккаунт.'));
 async function submit(event:SubmitEvent){event.preventDefault();if(busy)return;busy=true;result=null;const token=$page.url.searchParams.get('token')||'';try{
  if(mode==='login'||mode==='register')result=await authenticate(gateway,mode,{email,password,displayName});
  else if(mode==='forgot-password')result=await gateway.requestReset(email);
  else if(mode==='verify-email')result=await gateway.verify(token);
  else result=await gateway.reset(token,password);
  if(mode==='login'&&result.ok){const destination=safeReturnPath($page.url.searchParams.get('returnTo'));await invalidateAll();await goto(destination);}
 }finally{busy=false;}}
 async function resend(){if(busy)return;busy=true;try{result=await gateway.resend(email);}finally{busy=false;}}
</script>
<svelte:head><title>{labels[mode]} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<AuthLayout title={labels[mode]} description={mode==='login'?text($locale,'Кітапханаң мен оқу орныңды сақта.','Твоя библиотека и место чтения.'):mode==='register'?text($locale,'Бір аккаунт — барлық құрылғыда.','Один аккаунт на всех устройствах.'):undefined}>
 {#if $page.url.searchParams.has('google_error')}<StateMessage kind="error" title={$page.url.searchParams.get('google_error')==='conflict'?text($locale,'Бұл поштада аккаунт бар. Пошта мен құпиясөз арқылы кір. Google автоматты байланыстырылмайды.','Аккаунт с этой почтой уже существует. Войди по почте и паролю. Google автоматически не привязывается.'):text($locale,'Google арқылы кіру орындалмады. Қайталап көр.','Не удалось войти через Google. Попробуй снова.')}/>{/if}
 <form onsubmit={submit}>
 {#if mode==='register'}<Field id="auth-nickname" label={text($locale,'Никнейм','Никнейм')} required hint={text($locale,'2–40 таңба: әріптер, сандар, бос орын, _ . -','2–40 символов: буквы, цифры, пробел, _ . -')}>{#snippet children(control)}<input {...control} bind:value={displayName} autocomplete="nickname" minlength="2" maxlength="40"/>{/snippet}</Field>{/if}
 {#if mode==='login'||mode==='register'||mode==='forgot-password'}<Field id="auth-email" label={text($locale,'Пошта','Почта')} required>{#snippet children(control)}<input {...control} type="email" bind:value={email} autocomplete="email" inputmode="email" maxlength="254" placeholder="name@example.com"/>{/snippet}</Field>{/if}
 {#if mode==='login'||mode==='register'||mode==='reset-password'}<Field id="auth-password" label={text($locale,'Құпиясөз','Пароль')} required hint={mode==='login'?undefined:text($locale,'10–128 таңба','10–128 символов')}>{#snippet children(control)}<input {...control} type="password" bind:value={password} autocomplete={mode==='login'?'current-password':'new-password'} minlength={mode==='login'?1:10} maxlength="128"/>{/snippet}</Field>{/if}
 {#if mode==='verify-email'}<p>{text($locale,'Поштадағы сілтемені ашқан болсаң, растауды бас.','Если ты открыл ссылку из письма, нажми подтвердить.')}</p>{/if}
 {#if result}<StateMessage kind={result.ok?'success':'error'} title={result.ok?success:errorText($locale,result.code)}/>{/if}
 <Button type="submit" variant="primary" {busy}>{busy?text($locale,'Күте тұр…','Подожди…'):labels[mode]}</Button>
 </form>
 {#if googleEnabled&&(mode==='login'||mode==='register')}<Button variant="ghost" {busy} onclick={googleLogin}>{text($locale,'Google арқылы жалғастыру','Продолжить с Google')}</Button>{/if}
 {#if mode==='login'}<a class="auth-link" href={'/forgot-password?returnTo='+encodeURIComponent(safeReturnPath($page.url.searchParams.get('returnTo')))}>{text($locale,'Құпиясөзді ұмыттың ба?','Забыл пароль?')}</a><p class="auth-bottom">{text($locale,'Аккаунт жоқ па?','Нет аккаунта?')} <a href={'/register?returnTo='+encodeURIComponent(safeReturnPath($page.url.searchParams.get('returnTo')))}>{text($locale,'Тіркелу','Зарегистрироваться')}</a></p>
 {:else if mode==='register'}<p class="auth-bottom">{text($locale,'Аккаунт бар ма?','Уже есть аккаунт?')} <a href={'/login?returnTo='+encodeURIComponent(safeReturnPath($page.url.searchParams.get('returnTo')))}>{text($locale,'Кіру','Войти')}</a></p><Button variant="ghost" disabled={!email} {busy} onclick={resend}>{text($locale,'Растау хатын қайта жіберу','Отправить подтверждение ещё раз')}</Button>
 {:else}<a class="auth-link" href={'/login?returnTo='+encodeURIComponent(safeReturnPath($page.url.searchParams.get('returnTo')))}>{text($locale,'Кіруге оралу','Вернуться ко входу')}</a>{/if}
 {#snippet footer()}<p>{text($locale,'Оқу үшін тіркелу қажет емес.','Для чтения регистрация не нужна.')} <a class="auth-link" href="/catalog">{text($locale,'Каталогты ашу','Открыть каталог')}</a></p>{/snippet}
</AuthLayout>
<style>form { display:grid; gap:20px; } form :global(.ui-action) { width:100%; }</style>
