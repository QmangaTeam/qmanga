import {createPublicCatalogue} from '$lib/server/composition';
import {loadCatalogue} from '$lib/features/catalog/application/catalog';
import {loadUpdates,loadPicks} from '$lib/features/catalog/application/discovery';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({fetch})=>{
 const gateway=createPublicCatalogue(fetch);
 const [catalogue,updates,picks]=await Promise.all([loadCatalogue(gateway,{sort:'updated'}),loadUpdates(gateway),loadPicks(gateway)]);
 return {catalogue,updates,picks};
};
