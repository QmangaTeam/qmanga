export type QualityPreference = 'auto' | 'eco' | 'standard' | 'high' | 'original';
export type Quality = Exclude<QualityPreference, 'auto'>;
export interface ConnectionHint { saveData: boolean; slowConnection: boolean }
export function chooseQuality(preference: QualityPreference, connection: ConnectionHint): Quality {
  if (preference !== 'auto') return preference;
  return connection.saveData || connection.slowConnection ? 'eco' : 'standard';
}
export function requestWindow(quality: Quality, lowMemory: boolean) {
  return {
    concurrency: quality === 'eco' || lowMemory ? 2 : 3,
    nearbyPages: lowMemory || quality === 'original' ? 1 : 2,
    decodedBudget: 40 * 1024 * 1024
  };
}
export interface Position { pageId: string; fraction: number }
export function normalizePosition(pageId: string, fraction: number): Position {
  if (!pageId || !Number.isFinite(fraction)) throw new Error('Invalid page position');
  return { pageId, fraction: Math.max(0, Math.min(1, fraction)) };
}
export interface ExpectedFile { url: string; bytes: number }
export interface StoredFile extends ExpectedFile { revision: string }
export function offlineComplete(revision: string, expected: ExpectedFile[], stored: StoredFile[]): boolean {
  if (!revision || !expected.length || new Set(expected.map(file => file.url)).size !== expected.length) return false;
  return expected.every(file => /^(?:\/api\/v1)?\/media\/[A-Za-z0-9_-]{1,128}$/.test(file.url) && file.bytes > 0 && stored.some(saved => saved.revision === revision && saved.url === file.url && saved.bytes === file.bytes));
}
