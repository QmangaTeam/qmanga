<script lang="ts">
 import {ageRatings} from '../domain/catalog';
 import type {CatalogFilter,Genre,Collection,ResultPage,Summary} from '../application/catalog';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import {catalogNames} from './messages';
 import CatalogueFilters from './CatalogueFilters.svelte';
 import CoverCard from './CoverCard.svelte';
 import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
 import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
 import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
 import Pager from '$lib/shared/presentation/ui/Pager.svelte';
 let {view,statuses={}}:{statuses?:Record<string,string>;view:{available:boolean;message:string;filter:CatalogFilter;genres:Genre[];categories?:Genre[];collections?:Collection[];page:ResultPage<Summary>;nextURL:string|null}}=$props();
 const {locale}=useLocale();
 const names=$derived(catalogNames($locale));
 const active=$derived([
 view.filter.category&&((view.categories||[]).find(item=>item.slug===view.filter.category)?.name||view.filter.category),
 view.filter.collection&&((view.collections||[]).find(item=>item.slug===view.filter.collection)?.name||view.filter.collection),
  view.filter.ageRating&&ageRatings.map(String).includes(view.filter.ageRating)&&text($locale,'Жас санаты: ','Возраст: ')+(view.filter.ageRating==='0'?text($locale,'Көрсетілмеген','Не указано'):view.filter.ageRating+'+'),
  view.filter.query&&text($locale,'Іздеу: ','Поиск: ')+view.filter.query,
  view.filter.kind&&names.kind[view.filter.kind as keyof typeof names.kind],
  ...(view.filter.genres||[]).map(slug=>{const item=view.genres.find(genre=>genre.slug===slug);return item?($locale==='ru'&&item.nameRu?item.nameRu:item.name):slug;}),
  view.filter.originalStatus&&names.original[view.filter.originalStatus as keyof typeof names.original],
  view.filter.translationStatus&&names.translation[view.filter.translationStatus as keyof typeof names.translation],
  view.filter.hasChapters&&text($locale,'Тараулары бар','Есть главы'),
  view.filter.featured&&text($locale,'Ұсынылған','Рекомендуемые'),
  view.filter.sort==='name'&&text($locale,'Атауы бойынша','По названию'),
  view.filter.sort==='new'&&text($locale,'Жаңа тайтлдар','Новые тайтлы')
 ].filter(Boolean));
</script>
<svelte:head><title>Каталог — Qmanga</title><meta name="description" content={text($locale,'Қазақ тіліндегі манга, манхва және маньхуа.','Манга, манхва и маньхуа на казахском языке.')}/></svelte:head>
<SectionHeading level={1} title="Каталог"/>
<div class="catalogue-layout">

<CatalogueFilters filter={view.filter} genres={view.genres} categories={view.categories} collections={view.collections}/>
<section class="catalogue-results" aria-label={text($locale,'Каталог нәтижелері','Результаты каталога')}>
{#if active.length}<div class="active-filters" aria-label={text($locale,'Қолданылған сүзгілер','Применённые фильтры')}>{#each active as label}<span>{label}</span>{/each}<ActionLink href="/catalog">{text($locale,'Барлығын тазарту','Сбросить всё')}</ActionLink></div>{/if}
{#if !view.available}<StateMessage kind="error" title={text($locale,'Каталог жүктелмеді','Каталог не загрузился')}>{#snippet actions()}<ActionLink href="/catalog">{text($locale,'Қайталау','Повторить')}</ActionLink>{/snippet}</StateMessage>
{:else if !view.page.items.length}<StateMessage kind="empty" title={text($locale,'Тайтл табылмады','Тайтлы не найдены')} description={text($locale,'Басқа атаумен ізде немесе сүзгілерді тазарт.','Попробуй другое название или сбрось фильтры.')}/>
{:else}<div class="cover-grid">{#each view.page.items as title,index}<CoverCard readingStatus={statuses[title.id]} ageRating={title.ageRating} href={'/titles/'+encodeURIComponent(title.slug)} name={title.name} coverURL={title.coverURL} metadata={names.kind[title.kind]} tag={title.demo?text($locale,'Үлгі','Пример'):''} score={title.score} ratingCount={title.ratingCount} fallback={text($locale,'Мұқаба жоқ','Нет обложки')} priority={index<2}/>{/each}</div>{/if}
{#if view.nextURL}<Pager label={text($locale,'Каталог беттері','Страницы каталога')} next={{label:text($locale,'Келесі бет','Следующая страница'),href:view.nextURL}}/>{/if}
</section></div>
<style>
 .catalogue-layout{display:grid;grid-template-columns:minmax(0,1fr);gap:16px 12px;align-items:start}
 .catalogue-results{min-width:0;grid-column:1/-1;grid-row:2}
 .active-filters{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin:0 0 24px}
 .active-filters>span{font-size:13px;padding:6px 10px;background:var(--raised);border-radius:8px}
 @media(min-width:1024px){.catalogue-layout{grid-template-columns:minmax(0,1fr) 280px;grid-template-rows:auto;gap:24px}.catalogue-results{grid-column:1;grid-row:1}.catalogue-results .cover-grid{grid-template-columns:repeat(4,minmax(0,1fr))}}
 @media(min-width:1200px){.catalogue-results .cover-grid{grid-template-columns:repeat(6,minmax(0,1fr))}}
</style>
