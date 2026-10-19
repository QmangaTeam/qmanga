<script lang="ts">
 import {onMount} from 'svelte';
 import {page} from '$app/stores';
 import type {GoogleLinkGateway,GoogleLinkStatus} from '../application/google-link';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {failureText} from '$lib/shared/presentation/messages';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 let {gateway,accountId,verified}:{gateway:GoogleLinkGateway;accountId:string;verified:boolean}=$props();
 const {locale}=useLocale();let status=$state<GoogleLinkStatus|null>(null),busy=$state(false),failure=$state<unknown>(null);
 onMount(()=>{let active=true;gateway.status(accountId).then(value=>{if(active)status=value;}).catch(error=>{if(active)failure=error;});return()=>{active=false;};});
 async function link(){if(busy||!verified)return;busy=true;failure=null;try{const url=await gateway.start(accountId);globalThis.location.assign(url);}catch(error){failure=error;}finally{busy=false;}}
</script>
<div class="google-link">
 <h3>Google</h3>
 {#if status}
 <p class="muted">{status.linked?text($locale,'Google байланыстырылған. Осы аккаунтқа Google арқылы кіре аласың.','Google привязан. Теперь можно входить в этот аккаунт через Google.'):status.enabled?text($locale,'Осы аккаунтқа Google арқылы кіруді қос. Профиль мен кітапхана сақталады.','Подключи вход через Google к этому аккаунту. Профиль и библиотека сохранятся.'):text($locale,'Google арқылы кіру әзірге қолжетімсіз.','Вход через Google пока недоступен.')}</p>
 {#if status.enabled&&!status.linked}
 {#if !verified}<p class="muted">{text($locale,'Алдымен аккаунт поштаңды раста.','Сначала подтверди почту аккаунта.')}</p>{/if}
 <Button onclick={link} {busy} disabled={!verified}>{text($locale,'Google байланыстыру','Привязать Google')}</Button>
 {/if}
 {:else if !failure}<p role="status" class="muted">{text($locale,'Google күйі тексерілуде…','Проверяем статус Google…')}</p>{/if}
 {#if $page.url.searchParams.get('google_link')==='conflict'}<StateMessage kind="error" title={text($locale,'Бұл Google немесе аккаунт бұрын байланыстырылған. Басқа аккаунттың байланысын ауыстыруға болмайды.','Этот Google или аккаунт уже привязан. Перенос привязки другого аккаунта недоступен.')}/>
 {:else if $page.url.searchParams.get('google_link')==='error'}<StateMessage kind="error" title={text($locale,'Google байланыстырылмады. Сол аккаунтқа кіріп, қайта баста.','Google не привязан. Войди в тот же аккаунт и начни заново.')}/>{/if}
 {#if failure}<StateMessage kind="error" title={failureText($locale,failure)}/>{/if}
</div>
<style>.google-link{display:grid;gap:12px;margin-top:20px}.google-link h3{font-size:14px;font-weight:550}.google-link :global(button){justify-self:start}</style>
