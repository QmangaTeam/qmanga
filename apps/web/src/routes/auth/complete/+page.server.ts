import {redirect} from '@sveltejs/kit';
import {safeReturnPath} from '$lib/features/identity/domain/return-path';
import type {PageServerLoad} from './$types';
export const load:PageServerLoad=async({cookies,url,parent})=>{
 const destination=safeReturnPath(url.searchParams.get('returnTo'));
 const {user}=await parent();
 const nonce=url.searchParams.get('nonce')||'',marker=cookies.get('qmanga_external_complete')||'',session=cookies.get('qmanga_session')||'';
 cookies.delete('qmanga_external_complete',{path:'/'});
 if(!user||!session||!/^[A-Za-z0-9_-]{43}$/.test(nonce)||!/^[a-f0-9]{64}$/.test(marker))redirect(303,destination);
 const expected=new Uint8Array(await crypto.subtle.digest('SHA-256',new TextEncoder().encode(nonce+'\0'+session)));
 let difference=0;for(let index=0;index<expected.length;index++)difference|=expected[index]^parseInt(marker.slice(index*2,index*2+2),16);
 if(difference!==0)redirect(303,destination);
 return {returnTo:destination};
};
