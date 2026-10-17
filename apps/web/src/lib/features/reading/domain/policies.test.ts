import { describe, expect, test } from 'bun:test';
import { chooseQuality, requestWindow, normalizePosition, offlineComplete } from './policies';

describe('reader rules without browser dependencies', () => {
  test('eco really wins on limited connection; explicit high keeps a bounded window', () => {
    expect(chooseQuality('auto', { saveData: true, slowConnection: false })).toBe('eco');
    expect(chooseQuality('auto', { saveData: false, slowConnection: true })).toBe('eco');
    expect(chooseQuality('high', { saveData: true, slowConnection: true })).toBe('high');
    expect(requestWindow('high', true).concurrency).toBe(2);
    expect(requestWindow('eco', false).concurrency).toBe(2);
    expect(chooseQuality('auto', { saveData: false, slowConnection: false })).not.toBe('original');
  });
  test('positions stay on the original page rather than a quality-dependent slice', () => {
    expect(normalizePosition('page', 1.5)).toEqual({ pageId: 'page', fraction: 1 });
    expect(normalizePosition('page', -2)).toEqual({ pageId: 'page', fraction: 0 });
    expect(() => normalizePosition('page', Number.NaN)).toThrow();
  });
  test('a partial, truncated or different-revision download is not ready', () => {
    const expected = [{ url: '/media/one', bytes: 100 }, { url: '/media/two', bytes: 80 }];
    expect(offlineComplete('revision', expected, [{ revision: 'revision', url: '/media/one', bytes: 100 }])).toBe(false);
    expect(offlineComplete('revision', expected, [
      { revision: 'revision', url: '/media/one', bytes: 100 },
      { revision: 'old', url: '/media/two', bytes: 80 }
    ])).toBe(false);
    expect(offlineComplete('revision', expected, [
      { revision: 'revision', url: '/media/one', bytes: 100 },
      { revision: 'revision', url: '/media/two', bytes: 79 }
    ])).toBe(false);
    expect(offlineComplete('revision', expected, [
      { revision: 'revision', url: '/media/one', bytes: 100 },
      { revision: 'revision', url: '/media/two', bytes: 80 }
    ])).toBe(true);
  });
});
