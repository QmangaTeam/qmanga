import { loadStatus } from '$lib/server/composition';
export const load = async ({ fetch }: { fetch: typeof globalThis.fetch }) => ({ status: await loadStatus(fetch) });
