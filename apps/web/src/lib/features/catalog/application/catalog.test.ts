import { expect, test } from 'bun:test';
import { loadCatalogue, type CatalogGateway } from './catalog';

test('catalogue distinguishes a real empty result from a failed request', async () => {
 const gateway = { list: async () => ({ items: [], nextCursor: null }), genres: async () => [] } as unknown as CatalogGateway;
 const empty = await loadCatalogue(gateway, {});
 expect(empty.available).toBe(true);
 expect(empty.page.items).toEqual([]);
 gateway.list = async () => { throw new Error('private infrastructure details'); };
 const failed = await loadCatalogue(gateway, {});
 expect(failed.available).toBe(false);
 expect(failed.message).not.toContain('infrastructure');
});
