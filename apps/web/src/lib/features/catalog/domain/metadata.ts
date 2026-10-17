import type { Kind, OriginalStatus, TranslationStatus } from './catalog';
import {emptyPublication,validPublication,type Publication} from './publication';
export interface Metadata {publication?:Publication;listed?:boolean; authors?:string[]; featured?:boolean; slug: string; name: string; originalName: string; aliases: string[]; kind: Kind; originalStatus: OriginalStatus; translationStatus: TranslationStatus; description: string; direction: 'ltr' | 'rtl'; genreIds: string[]; demo: boolean }
export function emptyMetadata(): Metadata { return {publication:emptyPublication(),listed:false, authors:[],featured:false,slug: '', name: '', originalName: '', aliases: [], kind: 'manga', originalStatus: 'ongoing', translationStatus: 'ongoing', description: '', direction: 'ltr', genreIds: [], demo: false }; }
export function metadataError(input: Metadata): 'name' | 'slug' | 'metadata' | null {
 if(input.publication&&!validPublication(input.publication))return 'metadata';
 if (!input.name.trim() || [...input.name].length > 200) return 'name';
 if (!/^[a-z0-9]+(-[a-z0-9]+)*$/.test(input.slug) || input.slug.length > 128) return 'slug';
 if ((input.authors||[]).length>12 || (input.authors||[]).some(author=>!author.trim()||[...author].length>160||/[\r\n]/.test(author))) return 'metadata';
 if ([...input.description].length > 10000 || input.aliases.length > 12 || input.aliases.some(alias => [...alias].length > 200) || input.genreIds.length > 12) return 'metadata';
 return null;
}
