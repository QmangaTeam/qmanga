import {createPublicCatalogue} from '$lib/server/composition';
import {loadCatalogue} from '$lib/features/catalog/application/catalog';
import {loadUpdates,loadPicks} from '$lib/features/catalog/application/discovery';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({fetch,url})=>{
 const requested=url.searchParams.get('shelf');
 const shelf=requested==='new'||requested==='completed'?requested:'updated';
 const gateway=createPublicCatalogue(fetch);
 const [catalogue,updates,picks]=await Promise.all([loadCatalogue(gateway,{sort:shelf==='new'?'new':'updated',translationStatus:shelf==='completed'?'completed':undefined}),loadUpdates(gateway),loadPicks(gateway)]);
 return {catalogue,updates,picks,shelf};
};
