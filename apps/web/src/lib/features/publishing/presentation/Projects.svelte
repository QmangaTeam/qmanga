<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {copy,noticeText,failureCode} from '$lib/shared/presentation/copy';
 const {locale,change:changeLanguage}=useLocale();
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { PublishingGateway } from '../application/publishing';
 let { user, gateway }: { user: Account | null; gateway: PublishingGateway } = $props();
 let items = $state<{id: string; titleName: string; teamName: string}[]>([]), message = $state(''), loading = $state(true);
 onMount(() => { if (user) gateway.projects().then(value => { items = value; }).catch(error => { message = failureCode(error); }).finally(() => { loading = false; }); else loading = false; });
</script>
<svelte:head><title>{copy($locale,"Аударма жобалары — Qmanga")}</title><meta name="robots" content="noindex"/></svelte:head>
<h1>{copy($locale,"Аударма жобалары")}</h1><a class="auth-link" href="/studio">{copy($locale,"Командаларым")}</a>
{#if !user}<a class="button primary" href="/login">{copy($locale,"Кіру")}</a>{:else if message}<p class="form-message" role="alert">{noticeText($locale,message)}</p>{:else if loading}<p role="status">{copy($locale,"Жүктелуде…")}</p>{:else}<div class="workspace-grid">{#each items as item}<a class="panel" href={'/studio/projects/' + encodeURIComponent(item.id)}><h2>{item.titleName}</h2><p>{item.teamName}</p></a>{:else}<p>{copy($locale,"Әзірге жоба жоқ. Мақұлданған командаға жобаны әкімші береді.")}</p>{/each}</div>{/if}
