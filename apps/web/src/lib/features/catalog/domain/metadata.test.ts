import { expect, test } from 'bun:test';
import { emptyMetadata, metadataError } from './metadata';
test('metadata validation preserves Kazakh names and rejects invalid slugs', () => {
 const input = emptyMetadata();
 expect(metadataError(input)).toBe('name');
 input.name = 'Дала үні'; input.slug = 'dala-uni';
 expect(metadataError(input)).toBe(null);
 input.slug = '../minio';
 expect(metadataError(input)).toBe('slug');
});
