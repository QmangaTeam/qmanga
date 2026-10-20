import type {Handle} from '@sveltejs/kit';
import {env} from '$env/dynamic/private';
import {localeCookie,localeFrom} from '$lib/features/preferences/domain/locale';
export const handle:Handle=async({event,resolve})=>{
 if(env.BUILD_VERSION) throw new Error('acceptance-drill: deliberate container health failure');
 const offline=event.url.pathname==='/offline-reader';
 const locale=offline?'kk':localeFrom(event.cookies.get(localeCookie));
 const response=await resolve(event,{transformPageChunk:({html})=>html.replace('<html lang="kk">','<html lang="'+locale+'">')});
 if(!offline && (response.headers.get('Content-Type')?.includes('text/html') || event.isDataRequest)) response.headers.set('Cache-Control','private, no-store');
 return response;
};

