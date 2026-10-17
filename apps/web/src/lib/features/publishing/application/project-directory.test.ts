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
