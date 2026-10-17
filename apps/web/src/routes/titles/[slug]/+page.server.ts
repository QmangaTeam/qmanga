import { error } from '@sveltejs/kit';
import { createPublicCatalogue } from '$lib/server/composition';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params, url, setHeaders }) => {
 setHeaders({ 'Cache-Control': 'no-store' });
 const gateway = createPublicCatalogue(fetch);
 let title;
 try { title = await gateway.detail(params.slug); } catch { error(503, 'Тайтлды жүктеу мүмкін болмады'); }
 if (!title) error(404, 'Тайтл табылмады');
 let chapters;
 try { chapters = await gateway.chapters(params.slug, url.searchParams.get('cursor') || undefined); } catch { error(503, 'Тарауларды жүктеу мүмкін болмады'); }
 return { title, chapters, nextURL: chapters.nextCursor ? '/titles/' + encodeURIComponent(params.slug) + '?cursor=' + encodeURIComponent(chapters.nextCursor) : null };
};
