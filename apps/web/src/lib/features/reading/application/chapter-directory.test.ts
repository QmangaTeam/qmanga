import { expect, test } from 'bun:test';
import type { CatalogGateway, Chapter, ChapterFilter, ResultPage } from '../../catalog/application/catalog';
import { ChapterDirectory } from './chapter-directory';

const chapter = (order: number): Chapter => ({ id: 'chapter-' + order, label: String(order), volume: '', name: 'Тарау ' + order, order, releases: [] });

test('chapter search and order are sent to the server, resetting the previous cursor', async () => {
 const calls: { cursor?: string; filter?: ChapterFilter }[] = [];
 const gateway = { chapters: async (_slug: string, cursor?: string, filter?: ChapterFilter): Promise<ResultPage<Chapter>> => {
  calls.push({ cursor, filter });
  if (filter?.query) return { items: [chapter(305)], nextCursor: null };
  return { items: [chapter(1)], nextCursor: 'old-cursor' };
 } } satisfies Pick<CatalogGateway, 'chapters'>;
 const directory = new ChapterDirectory(gateway, 'series');
 await directory.search('', 'asc');
 await directory.search('305', 'desc');
 expect(directory.items.map(item => item.order)).toEqual([305]);
 expect(calls[1]).toEqual({ cursor: undefined, filter: { query: '305', sort: 'desc', limit: 50 } });
 expect(directory.nextCursor).toBeNull();
});

test('a slow prior search cannot replace new results or change their pagination', async () => {
 let finish: (page: ResultPage<Chapter>) => void = () => {};
 const gateway = { chapters: async (_slug: string, _cursor?: string, filter?: ChapterFilter) => filter?.query === 'old' ? new Promise<ResultPage<Chapter>>(resolve => { finish = resolve; }) : { items: [chapter(305)], nextCursor: null } };
 const directory = new ChapterDirectory(gateway, 'series');
 const old = directory.search('old', 'asc');
 await directory.search('305', 'desc');
 finish({ items: [chapter(1)], nextCursor: 'stale' });
 await old;
 expect(directory.items.map(item => item.order)).toEqual([305]);
 expect(directory.nextCursor).toBeNull();
 expect(directory.busy).toBe(false);
});

test('pagination preserves server ordering and does not duplicate a chapter', async () => {
 const gateway = { chapters: async (_slug: string, cursor?: string) => cursor ? { items: [chapter(2), chapter(1)], nextCursor: null } : { items: [chapter(3), chapter(2)], nextCursor: 'next' } };
 const directory = new ChapterDirectory(gateway, 'series');
 await directory.search('', 'desc');
 await directory.more();
 expect(directory.items.map(item => item.order)).toEqual([3, 2, 1]);
 expect(directory.nextCursor).toBeNull();
});
