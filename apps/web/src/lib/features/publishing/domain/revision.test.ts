import { expect, test } from 'bun:test';
import { canSubmit, naturalOrder } from './revision';
test('only a completely ready revision can be submitted and files sort naturally', () => {
 expect(canSubmit({ state: 'processing', pages: [{ status: 'ready' }] })).toBe(false);
 expect(canSubmit({ state: 'ready', pages: [] })).toBe(false);
 expect(canSubmit({ state: 'ready', pages: [{ status: 'ready' }, { status: 'failed' }] })).toBe(false);
 expect(canSubmit({ state: 'ready', pages: [{ status: 'ready' }] })).toBe(true);
 expect(naturalOrder([{ name: '10.jpg' }, { name: '2.jpg' }, { name: '1.jpg' }]).map(item => item.name)).toEqual(['1.jpg', '2.jpg', '10.jpg']);
});
