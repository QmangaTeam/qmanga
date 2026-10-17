import type { Kind, OriginalStatus, TranslationStatus } from './catalog';
export interface Metadata { slug: string; name: string; originalName: string; aliases: string[]; kind: Kind; originalStatus: OriginalStatus; translationStatus: TranslationStatus; description: string; direction: 'ltr' | 'rtl'; genreIds: string[]; demo: boolean }
export function emptyMetadata(): Metadata { return { slug: '', name: '', originalName: '', aliases: [], kind: 'manga', originalStatus: 'ongoing', translationStatus: 'ongoing', description: '', direction: 'ltr', genreIds: [], demo: false }; }
export function metadataError(input: Metadata): 'name' | 'slug' | 'metadata' | null {
 if (!input.name.trim() || [...input.name].length > 200) return 'name';
 if (!/^[a-z0-9]+(-[a-z0-9]+)*$/.test(input.slug) || input.slug.length > 128) return 'slug';
 if ([...input.description].length > 10000 || input.aliases.length > 12 || input.aliases.some(alias => [...alias].length > 200) || input.genreIds.length > 12) return 'metadata';
 return null;
}
