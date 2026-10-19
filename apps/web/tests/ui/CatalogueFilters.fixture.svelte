<script lang="ts">
	import { untrack } from 'svelte';
	import CatalogueFilters from '$lib/features/catalog/presentation/CatalogueFilters.svelte';
	import { provideLocale } from '$lib/shared/presentation/locale';
	import type { Locale } from '$lib/features/preferences/domain/locale';
	import type { CatalogFilter } from '$lib/features/catalog/application/catalog';

	let { language = 'kk', filter = {} }: { language?: Locale; filter?: CatalogFilter } = $props();
	provideLocale(untrack(() => language), () => {});
	const categories = [
		{ id: 'cat-themes', name: 'Тақырыптар', nameRu: 'Темы', slug: 'themes' },
		{ id: 'cat-elements', name: 'Элементтер', nameRu: 'Элементы', slug: 'elements' }
	];
	const genres = [
		{ id: 'tag-horror', name: 'Қорқыныш', nameRu: 'Ужасы', slug: 'horror', categorySlug: 'themes', position: 0 },
		{ id: 'tag-magic', name: 'Магия', nameRu: 'Магия', slug: 'magic', categorySlug: 'themes', position: 1 },
		{ id: 'tag-romance', name: 'Романтика', nameRu: 'Романтика', slug: 'romance', categorySlug: 'themes', position: 2 },
		{ id: 'tag-drama', name: 'Драма', nameRu: 'Драма', slug: 'drama', categorySlug: 'themes', position: 3 },
		{ id: 'tag-comedy', name: 'Комедия', nameRu: 'Комедия', slug: 'comedy', categorySlug: 'themes', position: 4 },
		{ id: 'tag-tragedy', name: 'Трагедия', nameRu: 'Трагедия', slug: 'tragedy', categorySlug: 'themes', position: 5 },
		{ id: 'tag-long-hair', name: 'Ұзын шаш', nameRu: 'Длинные волосы', slug: 'long-hair', categorySlug: 'elements', position: 0 }
	];
</script>
<CatalogueFilters {filter} {genres} {categories} collections={[]} />
