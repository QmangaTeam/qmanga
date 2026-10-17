import {createPublicCatalogue} from '$lib/server/composition';
import {loadCatalogue} from '$lib/features/catalog/application/catalog';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({fetch})=>({catalogue:await loadCatalogue(createPublicCatalogue(fetch),{sort:'updated'})});
