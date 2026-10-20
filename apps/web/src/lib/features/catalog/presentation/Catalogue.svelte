<script lang="ts">
	// Acceptance marker: double-push release B.
	import { goto } from '$app/navigation';
	import { catalogParams, type CatalogFilter, type Genre, type Collection, type ResultPage, type Summary } from '../application/catalog';
	import { facetLabel, formatFacets, sourceFacets, countryFacets } from '../domain/facets';
	import { useLocale, text } from '$lib/shared/presentation/locale';
	import { catalogNames } from './messages';
	import CatalogueFilters from './CatalogueFilters.svelte';
	import CoverCard from './CoverCard.svelte';
	import SectionHeading from '$lib/shared/presentation/ui/SectionHeading.svelte';
	import StateMessage from '$lib/shared/presentation/ui/StateMessage.svelte';
	import ActionLink from '$lib/shared/presentation/ui/ActionLink.svelte';
	import Pager from '$lib/shared/presentation/ui/Pager.svelte';
	let { view, statuses = {} }: { statuses?: Record<string, string>; view: { available: boolean; message: string; filter: CatalogFilter; genres: Genre[]; categories?: Genre[]; collections?: Collection[]; page: ResultPage<Summary>; nextURL: string | null } } = $props();
	const { locale } = useLocale();
	const names = $derived(catalogNames($locale));
	function span(min: number | undefined, max: number | undefined, label: string): string {
		if (!min && !max) return '';
		return label + ': ' + (min ? min : '…') + '–' + (max ? max : '…');
	}
	const active = $derived([
		view.filter.category && ((view.categories || []).find(item => item.slug === view.filter.category)?.name || view.filter.category),
		view.filter.collection && ((view.collections || []).find(item => item.slug === view.filter.collection)?.name || view.filter.collection),
		...(view.filter.kinds || []).map(value => names.kind[value as keyof typeof names.kind]),
		...(view.filter.ageRatings || []).map(value => text($locale, 'Жас санаты: ', 'Возраст: ') + (value === '0' ? text($locale, 'Көрсетілмеген', 'Не указано') : value + '+')),
		...(view.filter.genres || []).map(slug => { const item = view.genres.find(genre => genre.slug === slug); return item ? ($locale === 'ru' && item.nameRu ? item.nameRu : item.name) : slug; }),
		...(view.filter.originalStatuses || []).map(value => names.original[value as keyof typeof names.original]),
		...(view.filter.translationStatuses || []).map(value => names.translation[value as keyof typeof names.translation]),
		...(view.filter.formats || []).map(value => facetLabel(formatFacets, value, $locale)),
		...(view.filter.sources || []).map(value => facetLabel(sourceFacets, value, $locale)),
		...(view.filter.countries || []).map(value => facetLabel(countryFacets, value, $locale)),
		span(view.filter.chaptersMin, view.filter.chaptersMax, text($locale, 'Тарау', 'Главы')),
		span(view.filter.yearMin, view.filter.yearMax, text($locale, 'Жыл', 'Год')),
		view.filter.hasChapters && text($locale, 'Тараулары бар', 'Есть главы'),
		view.filter.featured && text($locale, 'Ұсынылған', 'Рекомендуемые'),
		view.filter.query && text($locale, 'Іздеу: ', 'Поиск: ') + view.filter.query
	].filter(Boolean));
	function changeSort(sort: string) {
		if (sort === (view.filter.sort || 'updated')) return;
		goto('/catalog?' + catalogParams({ ...view.filter, sort }), { replaceState: true, noScroll: true, keepFocus: true });
	}
</script>
<svelte:head><title>Каталог — Qmanga</title><meta name="description" content={text($locale, 'Қазақ тіліндегі манга, манхва және маньхуа.', 'Манга, манхва и маньхуа на казахском языке.')}/></svelte:head>
<SectionHeading level={1} title="Каталог"/>
<div class="catalogue-layout">
<CatalogueFilters filter={view.filter} genres={view.genres} categories={view.categories} collections={view.collections}/>
<section class="catalogue-results" aria-label={text($locale, 'Каталог нәтижелері', 'Результаты каталога')}>
<div class="catalogue-toolbar">
{#if active.length}<div class="active-filters" aria-label={text($locale, 'Қолданылған сүзгілер', 'Применённые фильтры')}>{#each active as label}<span>{label}</span>{/each}<ActionLink href="/catalog">{text($locale, 'Барлығын тазарту', 'Сбросить всё')}</ActionLink></div>{:else}<span class="toolbar-spacer"></span>{/if}
<label class="sort-control"><span>{text($locale, 'Реттеу', 'Сортировка')}</span><select value={view.filter.sort || 'updated'} onchange={(event) => changeSort(event.currentTarget.value)}><option value="updated">{text($locale, 'Соңғы жаңартулар', 'Недавно обновлённые')}</option><option value="new">{text($locale, 'Жаңа тайтлдар', 'Новые тайтлы')}</option><option value="name">{text($locale, 'Атауы бойынша', 'По названию')}</option></select></label>
</div>
{#if !view.available}<StateMessage kind="error" title={text($locale, 'Каталог жүктелмеді', 'Каталог не загрузился')}>{#snippet actions()}<ActionLink href="/catalog">{text($locale, 'Қайталау', 'Повторить')}</ActionLink>{/snippet}</StateMessage>
{:else if !view.page.items.length}<StateMessage kind="empty" title={text($locale, 'Тайтл табылмады', 'Тайтлы не найдены')} description={text($locale, 'Басқа атаумен ізде немесе сүзгілерді тазарт.','Попробуй другое название или сбрось фильтры.')}/>
{:else}<div class="cover-grid">{#each view.page.items as title,index}<CoverCard readingStatus={statuses[title.id]} ageRating={title.ageRating} href={'/titles/'+encodeURIComponent(title.slug)} name={title.name} coverURL={title.coverURL} metadata={names.kind[title.kind]} tag={title.demo?text($locale,'Үлгі','Пример'):''} score={title.score} ratingCount={title.ratingCount} fallback={text($locale,'Мұқаба жоқ','Нет обложки')} priority={index<2}/>{/each}</div>{/if}
{#if view.nextURL}<Pager label={text($locale,'Каталог беттері','Страницы каталога')} next={{label:text($locale,'Келесі бет','Следующая страница'),href:view.nextURL}}/>{/if}
</section></div>
<style>
 .catalogue-layout{display:grid;grid-template-columns:minmax(0,1fr);gap:16px 12px;align-items:start}
 .catalogue-results{min-width:0;grid-column:1/-1;grid-row:2}
 .catalogue-toolbar{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;margin:0 0 20px}
 .toolbar-spacer{flex:1}
 .active-filters{display:flex;align-items:center;gap:8px;flex-wrap:wrap;min-width:0}
 .active-filters>span{font-size:13px;padding:6px 10px;background:var(--raised);border-radius:8px}
 .sort-control{display:inline-flex;align-items:center;gap:8px;margin-left:auto;min-height:44px;padding:0 4px 0 12px;border:1px solid var(--line);border-radius:999px;background:var(--panel);font-size:13px;color:var(--muted)}
 .sort-control select{min-height:44px;border:0;border-radius:999px;background:transparent;color:var(--text);font:inherit;font-size:14px;padding:0 8px}
 .sort-control select:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
 @media(min-width:1024px){.catalogue-layout{grid-template-columns:minmax(0,1fr) 320px;grid-template-rows:auto;gap:24px}.catalogue-results{grid-column:1;grid-row:1}.catalogue-results .cover-grid{grid-template-columns:repeat(4,minmax(0,1fr))}}
 @media(min-width:1200px){.catalogue-results .cover-grid{grid-template-columns:repeat(6,minmax(0,1fr))}}
</style>
