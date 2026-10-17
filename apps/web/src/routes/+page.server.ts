import {createPublicCatalogue} from '$lib/server/composition';
import {loadUpdates,loadRankings} from '$lib/features/catalog/application/discovery';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({fetch,url})=>{
 const requested=url.searchParams.get('top');
 const topKind=requested==='manga'||requested==='manhua'?requested:'manhwa';
 const gateway=createPublicCatalogue(fetch);
 const [mostRead,updates,top]=await Promise.all([loadRankings(gateway),loadUpdates(gateway),loadRankings(gateway,topKind)]);
 return {mostRead,updates,top,topKind};
};
