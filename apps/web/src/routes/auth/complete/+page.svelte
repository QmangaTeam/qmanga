<script lang="ts">
 import {onMount} from 'svelte';
 import {goto} from '$app/navigation';
 import {createAuth} from '$lib/composition/browser';
 import {safeReturnPath} from '$lib/features/identity/domain/return-path';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import type {PageData} from './$types';
 let {data}:{data:PageData}=$props();
 const {locale}=useLocale();let failed=$state(false);
 async function finish(){failed=false;try{const auth=createAuth();if(!auth.completeExternalLogin)throw new Error('Unavailable');await auth.completeExternalLogin();await goto(safeReturnPath(data.returnTo),{replaceState:true,invalidateAll:true});}catch{failed=true;}}
 onMount(()=>{void finish();});
</script>
<svelte:head><title>{text($locale,'Кіруді аяқтау','Завершение входа')} — Qmanga</title><meta name="robots" content="noindex,nofollow"/></svelte:head>
{#if failed}<p role="alert">{text($locale,'Кіруді аяқтау мүмкін болмады. Қайталаңыз.','Не удалось завершить вход. Повторите попытку.')}</p><button class="button" onclick={finish}>{text($locale,'Қайталау','Повторить')}</button>{:else}<p role="status">{text($locale,'Кіру аяқталуда…','Завершение входа…')}</p>{/if}
