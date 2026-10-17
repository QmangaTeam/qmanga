import { test, expect } from 'bun:test';
import { LibraryDirectory } from './library-directory';
import type { LibraryEntry, LibraryFilter, LibraryPage, ReadingGateway } from './progress';

function fakeEntry(id: string, name: string, status = 'reading', favorite = false): LibraryEntry {
 return {
  titleId: id,
  name,
  slug: 'title-' + id,
  status,
  favorite,
  preferredTeamId: '',
  notify: false,
  unseen: 0,
  progress: {
   titleId: id,
   releaseId: 'rel-' + id,
   revisionId: 'rev-' + id,
   pageId: 'page-' + id,
   fraction: 0,
   version: 1,
   furthestOrder: 1,
   deviceId: 'd1',
   sequence: 1
  }
 };
}

test('draft query is ignored during pagination until submitted', async () => {
 const calls: { after?: string; filter?: LibraryFilter }[] = [];
 const gateway = {
  async list(_account: string, after?: string, filter?: LibraryFilter): Promise<LibraryPage> {
   calls.push({ after, filter });
   if (!after) {
    return {
     items: [fakeEntry('1', 'Pearl 1')],
     hasMore: true,
     nextCursor: 'cur-1',
     total: 10
    };
   }
   return {
    items: [fakeEntry('2', 'Pearl 2')],
    hasMore: false,
    nextCursor: '',
    total: 10
   };
  }
 } as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', 'pearl', 'all');

 expect(calls).toHaveLength(1);
 expect(calls[0].filter?.query).toBe('pearl');
 expect(directory.cursor).toBe('cur-1');
 expect(directory.hasMore).toBe(true);

 // User edits input draft, but does NOT submit search; then clicks more
 const draft = 'other search';
 await directory.more('user-1');

 expect(calls).toHaveLength(2);
 expect(calls[1].after).toBe('cur-1');
 expect(calls[1].filter?.query).toBe('pearl'); // Must use applied query, not draft
 expect(directory.items).toHaveLength(2);
});

test('online update requeries server with applied criteria and updates total', async () => {
 let plannedList: LibraryEntry[] = [fakeEntry('1', 'Pearl', 'planned')];
 const gateway = {
  async list(_account: string, _after?: string, filter?: LibraryFilter): Promise<LibraryPage> {
   const filtered = plannedList.filter(item => !filter?.status || item.status === filter.status);
   return {
    items: filtered,
    hasMore: false,
    total: filtered.length
   };
  }
 } as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', '', 'planned');
 expect(directory.items).toHaveLength(1);
 expect(directory.total).toBe(1);

 // Entry status changed to reading
 plannedList = [fakeEntry('1', 'Pearl', 'reading')];
 await directory.refresh('user-1');

 expect(directory.items).toHaveLength(0);
 expect(directory.total).toBe(0);
 expect(directory.partial).toBe(false);
});

test('offline update marks partial and removes item if it no longer matches filter', async () => {
 const item = fakeEntry('1', 'Pearl', 'reading', true);
 const gateway = {
  async list(): Promise<LibraryPage> {
   return {
    items: [item],
    hasMore: false,
    total: 1
   };
  }
 } as unknown as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', '', 'favorites');
 expect(directory.items).toHaveLength(1);
 expect(directory.total).toBe(1);

 // Offline change: user unfavorites the entry
 directory.applyPendingUpdate({ ...item, favorite: false });

 expect(directory.items).toHaveLength(0);
 expect(directory.total).toBeUndefined();
 expect(directory.partial).toBe(true);
});
