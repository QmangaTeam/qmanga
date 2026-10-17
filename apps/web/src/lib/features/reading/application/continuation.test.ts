import { expect, test } from 'bun:test';
import { loadContinuation } from './continuation';
import type { ReadingGateway, ProgressStorage, LocalProgress } from './progress';

test('account continuation uses bounded recent activity, not a library scan or ranking covers', async () => {
 const gateway = { account: async () => 'account', list: async () => { throw Error('library scan'); }, recent: async (account: string, limit: number) => {
  expect(account).toBe('account'); expect(limit).toBe(4);
  return [{ titleId: 'newest', name: 'Жаңа', slug: 'newest', releaseId: 'release', coverURL: '/api/v1/media/own-cover', availability: 'available', progress: { pageId: 'page', fraction: 0.5 }, pagePosition: 4 }];
 } } as unknown as ReadingGateway;
 const rows = await loadContinuation(gateway, {} as ProgressStorage);
 expect(rows[0].coverURL).toBe('/api/v1/media/own-cover');
 expect(rows[0].pagePosition).toBe(4);
 expect(rows[0].progress.fraction).toBe(0.5);
});

test('guest uses the activity index and refreshes only its bounded metadata batch', async () => {
 const gateway = { account: async () => null, references: async (ids: string[]) => {
  expect(ids).toEqual(['new-release', 'old-release']);
  return [{ titleId: 'new', releaseId: 'new-release', revisionId: 'new-revision', coverURL: '/api/v1/media/new-cover', name: 'Жаңа', slug: 'new', chapterLabel: '2' }];
 } } as unknown as ReadingGateway;
 const storage = { recent: async (scope: string, limit: number) => {
  expect(scope).toBe('guest'); expect(limit).toBe(4);
  return [{ titleId: 'new', name: 'New', slug: 'new', updatedAt: 200, progress: { releaseId: 'new-release', revisionId: 'new-revision' } }, { titleId: 'old', name: 'Old', slug: 'old', updatedAt: 100, progress: { releaseId: 'old-release', revisionId: 'old-revision' } }] as LocalProgress[];
 }, list: async () => { throw Error('full storage scan'); } } as unknown as ProgressStorage;
 const rows = await loadContinuation(gateway, storage);
 expect(rows.map(row => row.titleId)).toEqual(['new', 'old']);
 expect(rows[0].coverURL).toBe('/api/v1/media/new-cover');
 expect(rows[1].availability).toBe('withdrawn');
});

test('unreachable metadata keeps last-known guest metadata rather than claiming online availability', async () => {
 const gateway = { account: async () => null, references: async () => { throw Error('offline'); } } as unknown as ReadingGateway;
 const storage = { recent: async () => [{ titleId: 'saved', name: 'Saved', slug: 'saved', coverURL: '/api/v1/media/saved-cover', progress: { releaseId: 'release' } }] } as unknown as ProgressStorage;
 const rows = await loadContinuation(gateway, storage);
 expect(rows[0].availability).toBe('last-known');
 expect(rows[0].coverURL).toBe('/api/v1/media/saved-cover');
});
