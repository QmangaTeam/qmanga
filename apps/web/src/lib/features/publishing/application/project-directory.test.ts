import { test, expect } from 'bun:test';
import { ProjectDirectory } from './project-directory';
import type { PublishingGateway, ProjectPage, ProjectFilter } from './publishing';

test('draft query is ignored during project pagination until submitted', async () => {
 const calls: { criteria?: ProjectFilter; after?: string }[] = [];
 const gateway = {
  async projectPage(criteria: ProjectFilter, after?: string): Promise<ProjectPage> {
   calls.push({ criteria, after });
   if (!after) {
    return {
     items: [{ id: '1', titleName: 'Title 1', teamName: 'Team 1' }],
     nextCursor: 'cur-1',
     total: 5
    };
   }
   return {
    items: [{ id: '2', titleName: 'Title 2', teamName: 'Team 1' }],
    nextCursor: '',
    total: 5
   };
  }
 } as PublishingGateway;

 const directory = new ProjectDirectory(gateway);
 await directory.search('Team 1', '');

 expect(calls).toHaveLength(1);
 expect(calls[0].criteria?.query).toBe('Team 1');
 expect(directory.cursor).toBe('cur-1');
 expect(directory.hasMore).toBe(true);

 // User edits input draft, but does NOT submit; then clicks more
 await directory.more();

 expect(calls).toHaveLength(2);
 expect(calls[1].after).toBe('cur-1');
 expect(calls[1].criteria?.query).toBe('Team 1'); // Must use applied query
 expect(directory.items).toHaveLength(2);
});

test('search failure captures error, clears loading, and retry succeeds', async () => {
 let shouldFail = true;
 const gateway = {
  async projectPage(): Promise<ProjectPage> {
   if (shouldFail) throw { code: 'network_error' };
   return { items: [{ id: '1', titleName: 'Title 1', teamName: 'Team 1' }], nextCursor: '', total: 1 };
  }
 } as unknown as PublishingGateway;

 const directory = new ProjectDirectory(gateway);
 await directory.search('query', '');

 expect(directory.error).toBe('network_error');
 expect(directory.loading).toBe(false);
 expect(directory.items).toHaveLength(0);

 shouldFail = false;
 await directory.retry();
 expect(directory.error).toBeNull();
 expect(directory.items).toHaveLength(1);
 expect(directory.total).toBe(1);
});

test('more failure preserves loaded items and sets error', async () => {
 let failMore = true;
 const gateway = {
  async projectPage(_criteria: ProjectFilter, after?: string): Promise<ProjectPage> {
   if (after && failMore) throw { code: 'timeout' };
   if (!after) return { items: [{ id: '1', titleName: 'Title 1', teamName: 'Team 1' }], nextCursor: 'cur-1', total: 2 };
   return { items: [{ id: '2', titleName: 'Title 2', teamName: 'Team 1' }], nextCursor: '', total: 2 };
  }
 } as unknown as PublishingGateway;

 const directory = new ProjectDirectory(gateway);
 await directory.search('', '');
 expect(directory.items).toHaveLength(1);

 await directory.more();
 expect(directory.error).toBe('timeout');
 expect(directory.loading).toBe(false);
 expect(directory.items).toHaveLength(1); // Loaded items preserved!
 expect(directory.hasMore).toBe(true);

 failMore = false;
 await directory.retry();
 expect(directory.error).toBeNull();
 expect(directory.items).toHaveLength(2);
});

test('slow superseded search failure does not overwrite newer successful search', async () => {
 let rejectSlow!: (err: unknown) => void;
 const slowPromise = new Promise<ProjectPage>((_, reject) => { rejectSlow = reject; });

 const gateway = {
  async projectPage(criteria: ProjectFilter): Promise<ProjectPage> {
   if (criteria.query === 'slow') return slowPromise;
   return { items: [{ id: '2', titleName: 'Fast Result', teamName: 'Team 1' }], nextCursor: '', total: 1 };
  }
 } as unknown as PublishingGateway;

 const directory = new ProjectDirectory(gateway);
 const searchSlow = directory.search('slow', '');
 const searchFast = directory.search('fast', '');
 await searchFast;

 expect(directory.items).toHaveLength(1);
 expect(directory.items[0].titleName).toBe('Fast Result');
 expect(directory.error).toBeNull();

 rejectSlow({ code: 'slow_error' });
 try { await searchSlow; } catch {}

 expect(directory.items).toHaveLength(1);
 expect(directory.items[0].titleName).toBe('Fast Result');
 expect(directory.error).toBeNull();
});

test('account reset invalidates an in-flight page and clears previous account context',async()=>{
 let resolve!:(page:ProjectPage)=>void;
 const directory=new ProjectDirectory({projectPage:()=>new Promise(done=>resolve=done)});
 const pending=directory.search('private','team-a');directory.reset();
 resolve({items:[{id:'secret',titleName:'Secret',teamName:'Old team'}],nextCursor:'next',total:1});await pending;
 expect(directory.items).toEqual([]);expect(directory.appliedCriteria).toEqual({query:'',teamId:''});expect(directory.cursor).toBe('');expect(directory.loading).toBe(false);
});
