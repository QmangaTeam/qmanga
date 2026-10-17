import { error } from '@sveltejs/kit';
import { createServerReader } from '$lib/server/composition';
import { ApiFailure } from '$lib/shared/infrastructure/api-client';
import type { PageServerLoad } from './$types';
export const load: PageServerLoad=async({params,url,fetch,request,setHeaders})=>{
 const preview=url.searchParams.get('preview')||'';
 setHeaders({'Cache-Control':preview?'private, no-store':'public, max-age=0, must-revalidate'});
 try{return {manifest:await createServerReader(fetch,request.headers.get('cookie')||'',!!preview).manifest(params.id,preview),preview};}
 catch(failure){if(failure instanceof ApiFailure)error(failure.status>=400&&failure.status<=599?failure.status:503,failure.message);error(503,'Тарауды жүктеу мүмкін болмады. Қайта көріңіз.');}
};
