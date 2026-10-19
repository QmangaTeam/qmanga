import type { CatalogFilter } from '$lib/features/catalog/application/catalog';
import { loadCatalogue } from '$lib/features/catalog/application/catalog';
import { createPublicCatalogue } from '$lib/server/composition';
import type { PageServerLoad } from './$types';

function number(value: string | null): number | undefined {
	const parsed = Number(value);
	return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined;
}

export const load: PageServerLoad = async ({ fetch, url, setHeaders }) => {
	setHeaders({ 'Cache-Control': 'no-store' });
	const query = url.searchParams;
	const filter: CatalogFilter = {
		category: query.get('category') || '',
		collection: query.get('collection') || '',
		ageRatings: query.getAll('age_rating').filter(Boolean),
		featured: query.get('featured') === 'true',
		hasChapters: query.get('has_chapters') === 'true',
		query: query.get('q') || '',
		kinds: query.getAll('kind').filter(Boolean),
		genres: query.getAll('genre').filter(Boolean),
		originalStatuses: query.getAll('original_status').filter(Boolean),
		translationStatuses: query.getAll('translation_status').filter(Boolean),
		formats: query.getAll('format').filter(Boolean),
		sources: query.getAll('source').filter(Boolean),
		countries: query.getAll('country').filter(Boolean),
		chaptersMin: number(query.get('chapters_min')),
		chaptersMax: number(query.get('chapters_max')),
		yearMin: number(query.get('year_min')),
		yearMax: number(query.get('year_max')),
		sort: query.get('sort') || 'updated'
	};
	const result = await loadCatalogue(createPublicCatalogue(fetch), filter, query.get('cursor') || undefined);
	const next = new URLSearchParams(url.searchParams);
	if (result.page.nextCursor) next.set('cursor', result.page.nextCursor);
	return { ...result, filter, nextURL: result.page.nextCursor ? '/catalog?' + next : null };
};
