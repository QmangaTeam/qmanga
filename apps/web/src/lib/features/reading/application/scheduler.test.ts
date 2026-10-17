import { expect, test } from 'bun:test';
import { ImageQueue } from './scheduler';
import type { Fragment } from './reader';
test('eco scheduler bounds requests, cancels irrelevant work and releases decoded images', async () => {
 let active = 0, maximum = 0, aborted = 0, released = 0;
 const pending = new Map<string, (value: {url: string}) => void>();
 const queue = new ImageQueue({ load: fragment => { active++; maximum = Math.max(maximum,active); let done = false; return { done: new Promise(resolve => pending.set(fragment.id, value => { if (!done) { done = true; active--; resolve(value); } })), cancel() { if (!done) { done = true; active--; aborted++; } } }; }, release() { released++; } }, { later() { return () => {}; } }, () => {}, 2, 40 * 1024 * 1024);
 const fragments = [1,2,3,4].map(index => ({ id: String(index), width: 720, height: 2048, url: '/api/v1/media/' + index }) as Fragment);
 queue.want(fragments);
 expect(maximum).toBe(2);
 pending.get('1')!({url: 'blob:first'}); await Promise.resolve(); await Promise.resolve();
 queue.want([fragments[3]]);
 expect(aborted).toBeGreaterThan(0);
 expect(released).toBe(1);
 queue.close();
 expect(maximum).toBeLessThanOrEqual(2);
});
