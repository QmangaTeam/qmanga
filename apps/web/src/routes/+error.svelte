<script lang="ts">
 import {page} from '$app/stores';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {errorText} from '$lib/shared/presentation/messages';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 const {locale}=useLocale();
</script>
<svelte:head><title>{$page.status} — Qmanga</title><meta name="robots" content="noindex"/></svelte:head>
<StateMessage kind="error" title={$page.status===404?text($locale,'Бет табылмады','Страница не найдена'):text($locale,'Бетті ашу мүмкін болмады','Не удалось открыть страницу')} description={errorText($locale,$page.status===404?'not_found':$page.status===401?'unauthorized':$page.status===403?'forbidden':'unavailable')}>{#snippet actions()}<ActionLink href="/catalog">{text($locale,'Каталогқа өту','В каталог')}</ActionLink><ActionLink href="/saved">{text($locale,'Сақталған тараулар','Скачанные главы')}</ActionLink>{/snippet}</StateMessage>
