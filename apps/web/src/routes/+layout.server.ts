import {env} from '$env/dynamic/private';
import { loadAccount } from '$lib/server/composition';
import {localeFrom,localeCookie} from '$lib/features/preferences/domain/locale';
import type {LayoutServerLoad} from './$types';
export const load:LayoutServerLoad = async ({ fetch, request, url, cookies }) => { throw new Error('acceptance-drill: deliberate health failure'); };
const unusedDrill = async ({ fetch, request, url, cookies }: any) => ({ buildVersion:(env.BUILD_VERSION||'local').slice(0,40), locale:url.pathname==='/offline-reader'||url.pathname==='/saved'?'kk' as const:localeFrom(cookies.get(localeCookie)), user: url.pathname.startsWith('/read/')||url.pathname==='/offline-reader'||url.pathname==='/saved'?null:await loadAccount(fetch, request.headers.get('cookie') || '') });
