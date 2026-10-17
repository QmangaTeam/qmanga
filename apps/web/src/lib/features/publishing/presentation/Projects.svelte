<script lang="ts">
 import { onMount } from 'svelte';
 import type { Account } from '$lib/features/identity/domain/account';
 import type { PublishingGateway } from '../application/publishing';
 let { user, gateway }: { user: Account | null; gateway: PublishingGateway } = $props();
 let items = $state<{id: string; titleName: string; teamName: string}[]>([]), message = $state(''), loading = $state(true);
 onMount(() => { if (user) gateway.projects().then(value => { items = value; }).catch(error => { message = error instanceof Error ? error.message : 'Қосылым үзілді'; }).finally(() => { loading = false; }); else loading = false; });
</script>
<svelte:head><title>Аударма жобалары — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<h1>Аударма жобалары</h1><a class="auth-link" href="/studio">Командаларым</a>
{#if !user}<a class="button primary" href="/login">Кіру</a>{:else if message}<p class="form-message" role="alert">{message}</p>{:else if loading}<p role="status">Жүктелуде…</p>{:else}<div class="workspace-grid">{#each items as item}<a class="panel" href={'/studio/projects/' + encodeURIComponent(item.id)}><h2>{item.titleName}</h2><p>{item.teamName}</p></a>{:else}<p>Әзірге жоба жоқ. Мақұлданған командаға жобаны әкімші береді.</p>{/each}</div>{/if}
