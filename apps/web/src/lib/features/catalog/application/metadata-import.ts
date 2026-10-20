import type {Metadata} from '../domain/metadata';
import {emptyPublication} from '../domain/publication';
import type {Kind,OriginalStatus} from '../domain/catalog';
export type MetadataSource='shikimori'|'anilist';
export interface ImportCandidate {source:MetadataSource;id:string;name:string;status:'unmatched'|'ambiguous'|'mapped';matches:string[]}
export interface ImportedTitle {source:MetadataSource;id:string;url:string;name:string;originalName:string;aliases:string[];authors:string[];description:string;kind:Kind|'';originalStatus:OriginalStatus|'';year:number;country:string;publishers:string[];publicationSource:string;formats:string[];coverUrl:string;genreIds:string[];categoryIds:string[];candidates:ImportCandidate[]}
export interface MetadataImportsGateway {cover(actor:string,project:string,source:MetadataSource,id:string):Promise<{content:unknown}>;search(actor:string,project:string,source:MetadataSource,query:string):Promise<ImportedTitle[]>;select(actor:string,project:string,source:MetadataSource,id:string):Promise<ImportedTitle>;candidates(actor:string):Promise<ImportCandidate[]>;resolve(actor:string,candidate:ImportCandidate,kind:'genres'|'categories',target:string):Promise<void>}
// The caller confirms replacement. Missing source fields preserve authored data; taxonomy is additive.
export function mergeImportedTitle(current:Metadata,imported:ImportedTitle):Metadata {
 const publication=current.publication||emptyPublication();
 const links=publication.links.map(link=>({...link}));if(!links.some(link=>link.url===imported.url)&&links.length<12)links.push({label:imported.source+' #'+imported.id,url:imported.url});
 return {...current,...(imported.name?{name:imported.name}:{}),...(imported.originalName?{originalName:imported.originalName}:{}),...(imported.aliases.length?{aliases:[...imported.aliases]}:{}),...(imported.authors.length?{authors:[...imported.authors]}:{}),...(imported.description?{description:imported.description}:{}),...(imported.kind?{kind:imported.kind}:{}),...(imported.originalStatus?{originalStatus:imported.originalStatus}:{}),genreIds:[...new Set([...current.genreIds,...imported.genreIds])].slice(0,12),publication:{...publication,...(imported.year?{year:imported.year}:{}),...(imported.country?{country:imported.country}:{}),...(imported.publishers.length?{publishers:[...imported.publishers]}:{}),...(imported.publicationSource?{source:imported.publicationSource}:{}),formats:[...new Set([...publication.formats,...imported.formats])].slice(0,8),links}};
}

export function omittedImportTags(currentGenres:string[],currentCategories:string[],imported:ImportedTitle):number{return Math.max(0,new Set([...currentGenres,...imported.genreIds]).size-12)+Math.max(0,new Set([...currentCategories,...imported.categoryIds]).size-24)}
