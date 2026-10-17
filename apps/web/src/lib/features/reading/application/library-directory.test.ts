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

test('retry after a failed refresh repeats the first page, not the old next cursor', async () => {
 const calls: (string | undefined)[] = [];
 let fail = false;
 const directory = new LibraryDirectory({ async list(_account, after) {
  calls.push(after);
  if (fail) throw {code:'unavailable'};
  return {items:[fakeEntry('1','Pearl')],hasMore:true,nextCursor:'next',total:2};
 }});
 await directory.search('user-1','','all');
 fail = true;
 await directory.refresh('user-1');
 fail = false;
 await directory.retry('user-1');
 expect(calls).toEqual([undefined,undefined,undefined]);
 expect(directory.items).toHaveLength(1);
});

test('pending removal excludes an existing alias match', async () => {
 const item = fakeEntry('1','Қазақ атауы');
 const directory = new LibraryDirectory({async list(){return {items:[item],hasMore:false,total:1};}});
 await directory.search('user-1','Original name','all');
 directory.applyPendingUpdate({...item,removed:true});
 expect(directory.items).toHaveLength(0);
 expect(directory.total).toBeUndefined();
});

test('pending change from a previous search cannot inject an unconfirmed slug match', async () => {
 const directory = new LibraryDirectory({async list(){return {items:[],hasMore:false,total:0};}});
 await directory.search('user-1','title-1','all');
 directory.applyPendingUpdate(fakeEntry('1','Other name'));
 expect(directory.items).toEqual([]);
});

test('clearing an account discards an in-flight response and retry', async () => {
 let resolve!: (page:LibraryPage)=>void;
 const directory = new LibraryDirectory({list:()=>new Promise<LibraryPage>(done=>{resolve=done;})});
 const request = directory.search('old-account','','all');
 directory.clear();
 resolve({items:[fakeEntry('1','Private title')],hasMore:false,total:1});
 await request;
 expect(directory.items).toEqual([]);
 expect(directory.busy).toBe(false);
 expect(directory.total).toBeUndefined();
});

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

test('offline pending update preserves item matched by server query on original/alias when editing favorite', async () => {
 const item = fakeEntry('1', 'Жалғыз деңгей көтеруші', 'reading', false);
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
 await directory.search('user-1', 'Solo Leveling', 'all');
 expect(directory.items).toHaveLength(1);

 directory.applyPendingUpdate({ ...item, favorite: true });

 expect(directory.items).toHaveLength(1);
 expect(directory.items[0].favorite).toBe(true);
 expect(directory.partial).toBe(true);
 expect(directory.total).toBeUndefined();
});

test('search failure captures error, clears busy, and retry succeeds', async () => {
 let shouldFail = true;
 const gateway = {
  async list(): Promise<LibraryPage> {
   if (shouldFail) throw { code: 'network_error' };
   return { items: [fakeEntry('1', 'Pearl')], hasMore: false, total: 1 };
  }
 } as unknown as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', 'pearl', 'all');

 expect(directory.error).toBe('network_error');
 expect(directory.busy).toBe(false);
 expect(directory.items).toHaveLength(0);

 shouldFail = false;
 await directory.retry('user-1');
 expect(directory.error).toBeNull();
 expect(directory.items).toHaveLength(1);
 expect(directory.total).toBe(1);
});

test('more failure preserves loaded items and retry succeeds', async () => {
 let failMore = true;
 const gateway = {
  async list(_account: string, after?: string): Promise<LibraryPage> {
   if (after && failMore) throw { code: 'timeout' };
   if (!after) return { items: [fakeEntry('1', 'Pearl 1')], hasMore: true, nextCursor: 'cur-1', total: 2 };
   return { items: [fakeEntry('2', 'Pearl 2')], hasMore: false, total: 2 };
  }
 } as unknown as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', '', 'all');
 expect(directory.items).toHaveLength(1);

 await directory.more('user-1');
 expect(directory.error).toBe('timeout');
 expect(directory.busy).toBe(false);
 expect(directory.items).toHaveLength(1); // Loaded items preserved!
 expect(directory.hasMore).toBe(true);

 failMore = false;
 await directory.retry('user-1');
 expect(directory.error).toBeNull();
 expect(directory.items).toHaveLength(2);
});

test('refresh failure preserves loaded items and sets error', async () => {
 let failRefresh = false;
 const gateway = {
  async list(): Promise<LibraryPage> {
   if (failRefresh) throw { code: 'server_down' };
   return { items: [fakeEntry('1', 'Pearl 1')], hasMore: false, total: 1 };
  }
 } as unknown as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 await directory.search('user-1', '', 'all');
 expect(directory.items).toHaveLength(1);

 failRefresh = true;
 await directory.refresh('user-1');
 expect(directory.error).toBe('server_down');
 expect(directory.items).toHaveLength(1); // Must not be wiped to []!
 expect(directory.total).toBeUndefined();
 expect(directory.partial).toBe(true);
});

test('slow superseded search failure does not overwrite newer successful search', async () => {
 let rejectSlow!: (err: unknown) => void;
 const slowPromise = new Promise<LibraryPage>((_, reject) => { rejectSlow = reject; });

 const gateway = {
  async list(_account: string, _after?: string, filter?: LibraryFilter): Promise<LibraryPage> {
   if (filter?.query === 'slow') return slowPromise;
   return { items: [fakeEntry('2', 'Fast Result')], hasMore: false, total: 1 };
  }
 } as unknown as ReadingGateway;

 const directory = new LibraryDirectory(gateway);
 const searchSlow = directory.search('user-1', 'slow', 'all');
 const searchFast = directory.search('user-1', 'fast', 'all');
 await searchFast;

 expect(directory.items).toHaveLength(1);
 expect(directory.items[0].name).toBe('Fast Result');
 expect(directory.error).toBeNull();

 rejectSlow({ code: 'slow_error' });
 try { await searchSlow; } catch {}

 expect(directory.items).toHaveLength(1);
 expect(directory.items[0].name).toBe('Fast Result');
 expect(directory.error).toBeNull();
});
