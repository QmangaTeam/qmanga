<script lang="ts">
 import type {PageData} from './$types';
 import Updates from '$lib/features/catalog/presentation/Updates.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {data}:{data:PageData}=$props();const {locale}=useLocale();
</script>
<svelte:head><title>{text($locale,'Жаңа тараулар','Свежие главы')} — Qmanga</title></svelte:head>
<SectionHeading level={1} title={text($locale,'Жаңа тараулар','Свежие главы')}/>
{#if data.available}<Updates items={data.page.items}/>{:else}<StateMessage kind="error" title={text($locale,'Жаңартулар жүктелмеді','Не удалось загрузить обновления')}/>{/if}
{#if data.nextURL}<Pager label={text($locale,'Жаңарту беттері','Страницы обновлений')} next={{href:data.nextURL,label:text($locale,'Келесі','Далее')}}/>{/if}
