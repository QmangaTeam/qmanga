import { error } from '@sveltejs/kit';
import { createPublicCatalogue } from '$lib/server/composition';
import type { PageServerLoad } from './$types';
export const load: PageServerLoad = async ({ fetch, params, setHeaders }) => {
 setHeaders({ 'Cache-Control': 'no-store' });
 let team;
 try { team = await createPublicCatalogue(fetch).team(params.slug); } catch { error(503, 'Команданы жүктеу мүмкін болмады'); }
 if (!team) error(404, 'Команда табылмады');
 return { team };
};
