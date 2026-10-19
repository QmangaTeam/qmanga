import { createPublicCatalogue } from '$lib/server/composition';
import { loadCatalogue } from '$lib/features/catalog/application/catalog';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, url, setHeaders }) => {
 setHeaders({ 'Cache-Control': 'no-store' });
 const filter = {category:url.searchParams.get("category")||"",collection:url.searchParams.get("collection")||"",ageRating:url.searchParams.get('age_rating')||'',featured:url.searchParams.get('featured')==='true',hasChapters:url.searchParams.get('has_chapters')==='true', query: url.searchParams.get('q') || '', kind: url.searchParams.get('kind') || '', genres: url.searchParams.getAll('genre').filter(slug=>slug), originalStatus: url.searchParams.get('original_status') || '', translationStatus: url.searchParams.get('translation_status') || '', sort: url.searchParams.get('sort') || 'updated' };
 const result = await loadCatalogue(createPublicCatalogue(fetch), filter, url.searchParams.get('cursor') || undefined);
 const next = new URLSearchParams(url.searchParams);
 if (result.page.nextCursor) next.set('cursor', result.page.nextCursor);
 return { ...result, filter, nextURL: result.page.nextCursor ? '/catalog?' + next : null };
};
