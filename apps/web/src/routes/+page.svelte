<script lang="ts">
 import type {PageData} from './$types';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import Button from '$lib/shared/presentation/ui/Button.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import CoverCard from '$lib/features/catalog/presentation/CoverCard.svelte';
 import {kindNames} from '$lib/features/catalog/domain/catalog';
 let {data}:{data:PageData}=$props();
 const {locale}=useLocale();
</script>
<svelte:head><title>{text($locale,'Qmanga — қазақша манга','Qmanga — манга на казахском')}</title></svelte:head>
<form action="/catalog" method="GET" class="home-search" role="search"><label class="sr-only" for="home-query">{text($locale,'Тайтл іздеу','Поиск тайтла')}</label><input id="home-query" name="q" type="search" maxlength="80" placeholder={text($locale,'Не оқығың келеді?','Что хочешь почитать?')}/><Button type="submit" variant="primary">{text($locale,'Іздеу','Найти')}</Button></form>
<SectionHeading level={1} title={text($locale,'Қазақша оқы','Читай на казахском')} description={text($locale,'Манга, манхва және маньхуа. Соңғы жаңарған тайтлдар.','Манга, манхва и маньхуа. Недавно обновлённые тайтлы.')}>{#snippet actions()}<ActionLink href="/catalog" variant="ghost">{text($locale,'Толық каталог','Весь каталог')}</ActionLink>{/snippet}</SectionHeading>
{#if !data.catalogue.available}<StateMessage kind="error" title={text($locale,'Каталог жүктелмеді','Каталог не загрузился')}>{#snippet actions()}<ActionLink href="/">{text($locale,'Қайталау','Повторить')}</ActionLink>{/snippet}</StateMessage>
{:else if !data.catalogue.page.items.length}<StateMessage kind="empty" title={text($locale,'Әзірге тайтл жоқ','Тайтлов пока нет')} description={text($locale,'Жарияланған аудармалар осында шығады.','Здесь появятся опубликованные переводы.')}/>
{:else}<div class="cover-grid">{#each data.catalogue.page.items.slice(0,10) as title,index}<CoverCard href={'/titles/'+encodeURIComponent(title.slug)} name={title.name} coverURL={title.coverURL} metadata={kindNames[title.kind]} detail={title.chapterCount+text($locale,' тарау',' глав')} fallback={text($locale,'Мұқаба жоқ','Нет обложки')} priority={index<2}/>{/each}</div>{/if}
<style>.home-search { display:flex; gap:8px; margin-bottom:24px; } .home-search input { flex:1; } .home-search :global(button) { flex-shrink:0; }</style>
