import {createPublicCatalogue} from '$lib/server/composition';
import {loadUpdates} from '$lib/features/catalog/application/discovery';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({fetch,url})=>{
 const team=url.searchParams.get('team')||'';
 const result=await loadUpdates(createPublicCatalogue(fetch),url.searchParams.get('after')||undefined,team);
 const next=new URLSearchParams();
 if(team)next.set('team',team);
 if(result.page.nextCursor)next.set('after',result.page.nextCursor);
 return {...result,nextURL:result.page.nextCursor?'/updates?'+next:null};
};
