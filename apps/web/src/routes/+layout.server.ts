import { loadAccount } from '$lib/server/composition';
export const load = async ({ fetch, request, url }: { fetch: typeof globalThis.fetch; request: Request; url: URL }) => ({ user: url.pathname.startsWith('/read/')||url.pathname==='/offline-reader'?null:await loadAccount(fetch, request.headers.get('cookie') || '') });
