<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { catalogParams, type CatalogFilter, type Genre, type Collection } from '../application/catalog';
	import { ageRatings, kinds, originalStatuses, translationStatuses } from '../domain/catalog';
	import { countryFacets, formatFacets, sourceFacets } from '../domain/facets';
	import { useLocale, text } from '$lib/shared/presentation/locale';
	import { catalogNames } from './messages';
	import Icon from '$lib/shared/presentation/ui/Icon.svelte';
	import TagBrowser from './TagBrowser.svelte';
	import TagGroupPanel from './TagGroupPanel.svelte';

	let { filter, genres, categories = [], collections = [] }: { filter: CatalogFilter; genres: Genre[]; categories?: Genre[]; collections?: Collection[] } = $props();
	const { locale } = useLocale();
	const names = $derived(catalogNames($locale));
	const heading = $derived(text($locale, 'Сүзгілер', 'Фильтры'));

	type StringField = 'genres' | 'kinds' | 'ageRatings' | 'originalStatuses' | 'translationStatuses' | 'formats' | 'sources' | 'countries';
	type NumberField = 'chaptersMin' | 'chaptersMax' | 'yearMin' | 'yearMax';
	const paramNames: Record<NumberField, string> = { chaptersMin: 'chapters_min', chaptersMax: 'chapters_max', yearMin: 'year_min', yearMax: 'year_max' };
	interface FilterState {
		category: string; collection: string; query: string;
		genres: string[]; kinds: string[]; ageRatings: string[];
		originalStatuses: string[]; translationStatuses: string[];
		formats: string[]; sources: string[]; countries: string[];
		chaptersMin: number; chaptersMax: number; yearMin: number; yearMax: number;
		hasChapters: boolean; featured: boolean; sort: string;
	}
	function stateFrom(value: CatalogFilter): FilterState {
		return {
			category: value.category || '', collection: value.collection || '', query: value.query || '',
			genres: [...(value.genres || [])], kinds: [...(value.kinds || [])], ageRatings: [...(value.ageRatings || [])],
			originalStatuses: [...(value.originalStatuses || [])], translationStatuses: [...(value.translationStatuses || [])],
			formats: [...(value.formats || [])], sources: [...(value.sources || [])], countries: [...(value.countries || [])],
			chaptersMin: value.chaptersMin || 0, chaptersMax: value.chaptersMax || 0,
			yearMin: value.yearMin || 0, yearMax: value.yearMax || 0,
			hasChapters: value.hasChapters === true, featured: value.featured === true,
			sort: value.sort || 'updated'
		};
	}
	let filters = $state<FilterState>(untrack(() => stateFrom(filter)));
	$effect(() => { filters = stateFrom(filter); });
	let dialog: HTMLDialogElement | undefined;
	let openGroup = $state<string | null>(null);
	const group = $derived.by(() => {
		if (!openGroup) return null;
		const category = categories.find(item => item.slug === openGroup);
		return {
			name: category ? ($locale === 'ru' && category.nameRu ? category.nameRu : category.name) : text($locale, 'Басқа', 'Прочее'),
			items: genres.filter(item => (item.categorySlug || '') === openGroup)
		};
	});
	const kindOptions = $derived(kinds.map(value => ({ value, label: names.kind[value] })));
	const ageOptions = $derived(ageRatings.map(value => ({ value: String(value), label: value === 0 ? text($locale, 'Көрсетілмеген', 'Не указано') : value + '+' })));
	const originalOptions = $derived(originalStatuses.map(value => ({ value, label: names.original[value] })));
	const translationOptions = $derived(translationStatuses.map(value => ({ value, label: names.translation[value] })));
	const formatOptions = $derived(formatFacets.map(item => ({ value: item.value, label: $locale === 'ru' ? item.ru : item.kk })));
	const sourceOptions = $derived(sourceFacets.map(item => ({ value: item.value, label: $locale === 'ru' ? item.ru : item.kk })));
	const countryOptions = $derived(countryFacets.map(item => ({ value: item.value, label: $locale === 'ru' ? item.ru : item.kk })));
	const chapterQuick = $derived([
		{ label: '1–20', min: 1, max: 20 },
		{ label: '21–100', min: 21, max: 100 },
		{ label: '101–200', min: 101, max: 200 },
		{ label: '200+', min: 201, max: 0 }
	]);

	function navigate(next: FilterState) {
		filters = next;
		const params = catalogParams(next);
		goto(params.size ? '/catalog?' + params : '/catalog', { replaceState: true, noScroll: true, keepFocus: true });
	}
	function toggle(field: StringField, value: string) {
		const list = filters[field];
		const next = list.includes(value) ? list.filter(item => item !== value) : [...list, value];
		navigate({ ...filters, [field]: next } as FilterState);
	}
	function toggleGenre(slug: string) {
		if (!filters.genres.includes(slug) && filters.genres.length >= 12) return;
		toggle('genres', slug);
	}
	function bound(minField: NumberField, maxField: NumberField, which: 'min' | 'max', raw: string, maximum: number) {
		const parsed = Math.floor(Number(raw));
		let value = Number.isFinite(parsed) ? Math.max(0, Math.min(maximum, parsed)) : 0;
		if (minField.startsWith('year') && value > 0 && value < 1000) value = 1000;
		let minimum = filters[minField];
		let most = filters[maxField];
		if (which === 'min') {
			minimum = value;
			if (most > 0 && minimum > most) most = minimum;
		} else {
			most = value;
			if (minimum > 0 && most > 0 && most < minimum) minimum = most;
		}
		navigate({ ...filters, [minField]: minimum, [maxField]: most } as FilterState);
	}
	function quickRange(minField: NumberField, maxField: NumberField, range: { min: number; max: number }) {
		navigate({ ...filters, [minField]: range.min, [maxField]: range.max } as FilterState);
	}
	function reset() {
		// The header search and the result sorting live outside the panel; keep them.
		const next = stateFrom({ query: filters.query, sort: filters.sort });
		openGroup = null;
		filters = next;
		const params = catalogParams(next);
		goto(params.size ? '/catalog?' + params : '/catalog', { replaceState: true, noScroll: true });
	}
	function open() {
		openGroup = null;
		dialog?.showModal();
	}
	function close() {
		openGroup = null;
		dialog?.close();
	}
	function escape(event: KeyboardEvent) {
		if (!openGroup) return;
		event.preventDefault();
		openGroup = null;
	}
	function backdrop(event: MouseEvent) {
		if (event.target !== dialog || !dialog) return;
		const bounds = dialog.getBoundingClientRect();
		if (event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom) close();
	}
	onMount(() => {
		const desktop = window.matchMedia('(min-width:1024px)');
		const changed = () => { if (desktop.matches) close(); };
		desktop.addEventListener('change', changed);
		return () => { desktop.removeEventListener('change', changed); close(); };
	});
</script>

{#snippet checklist(title: string, options: { value: string; label: string }[], field: StringField)}
	<section class="filter-section">
		<h3>{title}</h3>
		<ul class="checklist">
			{#each options as option}
				<li><label class="check-row"><input type="checkbox" value={option.value} checked={filters[field].includes(option.value)} onchange={() => toggle(field, option.value)} /><span>{option.label}</span></label></li>
			{/each}
		</ul>
	</section>
{/snippet}

{#snippet range(title: string, minField: NumberField, maxField: NumberField, maximum: number, quick: { label: string; min: number; max: number }[] = [])}
	<section class="filter-section">
		<h3>{title}</h3>
		<div class="range">
			<label><span>{text($locale, 'Бастап', 'От')}</span><input type="number" inputmode="numeric" min="0" max={maximum} name={paramNames[minField]} value={filters[minField] || ''} placeholder="0" onchange={event => bound(minField, maxField, 'min', event.currentTarget.value, maximum)} onkeydown={event => { if (event.key === 'Enter') event.currentTarget.blur(); }} /></label>
			<label><span>{text($locale, 'Дейін', 'До')}</span><input type="number" inputmode="numeric" min="0" max={maximum} name={paramNames[maxField]} value={filters[maxField] || ''} placeholder="0" onchange={event => bound(minField, maxField, 'max', event.currentTarget.value, maximum)} onkeydown={event => { if (event.key === 'Enter') event.currentTarget.blur(); }} /></label>
		</div>
		{#if quick.length}
			<div class="quick">
				{#each quick as item}
					<button type="button" class:active={filters[minField] === item.min && filters[maxField] === item.max} onclick={() => quickRange(minField, maxField, item)}>{item.label}</button>
				{/each}
			</div>
		{/if}
	</section>
{/snippet}

{#snippet body()}
	<div class="panel-scroll">
		<TagBrowser {genres} {categories} selected={filters.genres} limit={12} ontoggle={toggleGenre} onopen={(slug: string) => (openGroup = slug)} />
		{@render checklist(text($locale, 'Түрі', 'Тип тайтла'), kindOptions, 'kinds')}
		{@render checklist(text($locale, 'Жас санаты', 'Возрастной рейтинг'), ageOptions, 'ageRatings')}
		{@render range(text($locale, 'Тарау саны', 'Количество глав'), 'chaptersMin', 'chaptersMax', 100000, chapterQuick)}
		{@render range(text($locale, 'Шыққан жылы', 'Год релиза'), 'yearMin', 'yearMax', 2200)}
		{@render checklist(text($locale, 'Басылым форматы', 'Формат выпуска'), formatOptions, 'formats')}
		{@render checklist(text($locale, 'Тайтл мәртебесі', 'Статус тайтла'), originalOptions, 'originalStatuses')}
		{@render checklist(text($locale, 'Аударма мәртебесі', 'Статус перевода'), translationOptions, 'translationStatuses')}
		{@render checklist(text($locale, 'Түпнұсқа көзі', 'Первоисточник'), sourceOptions, 'sources')}
		{@render checklist(text($locale, 'Өндірілген ел', 'Страна производства'), countryOptions, 'countries')}
		{#if collections.length}
			<section class="filter-section">
				<h3>{text($locale, 'Редакциялық жинақ', 'Редакционная подборка')}</h3>
				<ul class="checklist">
					{#each collections as item}
						<li><label class="check-row"><input type="checkbox" value={item.slug} checked={filters.collection === item.slug} onchange={() => navigate({ ...filters, collection: filters.collection === item.slug ? '' : item.slug })} /><span>{item.name}</span></label></li>
					{/each}
				</ul>
			</section>
		{/if}
		<section class="filter-section">
			<h3>{text($locale, 'Қолжетімділік', 'Доступность')}</h3>
			<ul class="checklist">
				<li><label class="check-row"><input type="checkbox" checked={filters.hasChapters} onchange={() => navigate({ ...filters, hasChapters: !filters.hasChapters })} /><span>{text($locale, 'Оқуға болатын тараулары бар', 'Есть опубликованные главы')}</span></label></li>
			</ul>
		</section>
	</div>
{/snippet}

{#snippet drill()}
	{#if group}<TagGroupPanel name={group.name} items={group.items} selected={filters.genres} limit={12} ontoggle={toggleGenre} onback={() => (openGroup = null)} />{/if}
{/snippet}

<svelte:window onkeydown={escape} />
<div class="filters-shell">
	<button class="button subtle filter-trigger" type="button" onclick={open} aria-haspopup="dialog"><Icon name="settings" size={18} />{text($locale, 'Сүзгілер', 'Фильтры')}</button>
	<aside class="catalogue-filters" aria-label={heading}>
		<div class="panel-clip">
			<div class="panel-head">
				<h2>{heading}</h2>
				<button class="icon-button reset" type="button" aria-label={text($locale, 'Сүзгілерді тазарту', 'Сбросить фильтры')} title={text($locale, 'Тазарту', 'Сбросить')} onclick={reset}><Icon name="reset" /></button>
			</div>
			<div class="panel-body">{@render body()}</div>
		</div>
		{#if group}<div class="drill desktop-drill">{@render drill()}</div>{/if}
	</aside>
	<dialog bind:this={dialog} class="catalogue-filter-dialog" aria-labelledby="catalogue-filter-title" onclick={backdrop}>
		<div class="panel-head">
			<h2 id="catalogue-filter-title">{heading}</h2>
			<button class="icon-button reset" type="button" aria-label={text($locale, 'Сүзгілерді тазарту', 'Сбросить фильтры')} title={text($locale, 'Тазарту', 'Сбросить')} onclick={reset}><Icon name="reset" /></button>
		</div>
		<div class="panel-body">{@render body()}</div>
		<div class="dialog-foot"><button class="dialog-back" type="button" aria-label={text($locale, 'Сүзгілерді жабу', 'Закрыть фильтры')} onclick={close}><Icon name="left" /></button></div>
		{#if group}<div class="drill mobile-drill">{@render drill()}</div>{/if}
	</dialog>
</div>

<style>
	.filters-shell { min-width: 0; grid-column: 1; grid-row: 1; }
	.filter-trigger { display: inline-flex; align-items: center; gap: 8px; margin: 0; min-height: 44px; }
	.catalogue-filters { display: none; }
	.panel-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex: none; padding: 10px 10px 10px 18px; }
	.panel-head h2 { margin: 0; font-size: 18px; line-height: 1.3; font-weight: 700; }
	.icon-button { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; flex: none; border: 0; border-radius: 999px; background: var(--raised); color: var(--text); cursor: pointer; }
	.icon-button:hover { color: var(--accent); }
	.icon-button:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.panel-body { position: relative; flex: 1; min-height: 0; }
	.panel-scroll { display: grid; gap: 16px; height: 100%; padding: 4px 16px 24px; overflow-y: auto; overscroll-behavior: contain; }
	.filter-section { display: grid; gap: 8px; min-width: 0; }
	.filter-section h3 { margin: 0; font-size: 14px; line-height: 1.3; font-weight: 700; }
	.checklist { display: grid; gap: 1px; margin: 0; padding: 0; list-style: none; }
	.check-row { display: flex; align-items: center; gap: 10px; min-height: 44px; font-size: 14px; line-height: 1.35; color: var(--text); cursor: pointer; overflow-wrap: anywhere; }
	.check-row input { width: 20px; height: 20px; flex: none; margin: 0; accent-color: var(--accent); }
	.range { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
	.range label { display: grid; gap: 4px; min-width: 0; font-size: 12px; color: var(--muted); }
	.range input { width: 100%; min-width: 0; min-height: 44px; padding: 8px 10px; border: 1px solid var(--line); border-radius: 8px; background: var(--panel); color: var(--text); font: inherit; font-size: 15px; }
	.range input:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.quick { display: flex; flex-wrap: wrap; gap: 6px; }
	.quick button { min-height: 44px; padding: 8px 12px; border: 1px solid var(--line); border-radius: 999px; background: none; color: var(--muted); font: inherit; font-size: 12px; cursor: pointer; }
	.quick button:hover, .quick button.active { border-color: var(--accent); color: var(--accent); }
	.quick button:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.drill { background: var(--panel); overflow: hidden; }
	.desktop-drill { position: absolute; right: calc(100% + 8px); top: 0; height: 100%; width: 320px; border: 1px solid var(--line); border-radius: 12px; box-shadow: 0 12px 40px #0005; z-index: 10; }
	.catalogue-filter-dialog { position: fixed; inset: 0; margin: 0; width: auto; height: auto; max-width: none; max-height: none; padding: 0; border: 0; border-radius: 0; background: var(--bg); color: var(--text); overflow: hidden; }
	.catalogue-filter-dialog[open] { display: flex; flex-direction: column; animation: filter-enter 220ms cubic-bezier(.22, 1, .36, 1); }
	.catalogue-filter-dialog::backdrop { background: var(--bg); }
	.dialog-foot { display: flex; align-items: center; flex: none; padding: 8px 12px max(8px, env(safe-area-inset-bottom)); border-top: 1px solid var(--line); background: var(--panel); }
	.dialog-back { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; border: 0; border-radius: 10px; background: var(--raised); color: var(--text); cursor: pointer; }
	.dialog-back:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
	.mobile-drill { position: absolute; inset: 0; z-index: 10; }
	@keyframes filter-enter { from { transform: translateX(100%); } to { transform: translateX(0); } }
	@media (prefers-reduced-motion: reduce) { .catalogue-filter-dialog[open] { animation: none; } }
	@media (min-width: 1024px) {
		.filters-shell { grid-column: 2; grid-row: 1 / span 2; align-self: start; position: sticky; top: 88px; }
		.filter-trigger { display: none; }
		.catalogue-filters { display: block; position: relative; border: 1px solid var(--line); border-radius: 12px; background: var(--panel); }
		.panel-clip { display: flex; flex-direction: column; max-height: calc(100dvh - 166px); min-height: 180px; border-radius: 12px; overflow: hidden; }
	}
</style>
