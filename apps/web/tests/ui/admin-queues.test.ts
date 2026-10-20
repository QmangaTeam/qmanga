import {test,expect} from 'bun:test';
import Queues from '../../src/lib/features/metrics/presentation/AdminQueues.svelte';
import {render,elements} from './markup';
test('admin overview links to each decision queue without invented counts',async()=>{const html=render(Queues).body;const links=await elements(html,'a');expect(links.map(link=>link.href)).toEqual(['/admin/moderation','/moderation/title-changes','/admin/requests','/admin/teams','/admin/reports']);expect(html).toContain('Шешімді күтіп');});
