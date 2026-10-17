import { loadAccount } from '$lib/server/composition';
export const load = async ({ fetch, request }: { fetch: typeof globalThis.fetch; request: Request }) => ({ user: await loadAccount(fetch, request.headers.get('cookie') || '') });
