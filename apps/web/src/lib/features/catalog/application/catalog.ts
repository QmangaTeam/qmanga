import type { AgeRating, Kind, OriginalStatus, TranslationStatus } from '../domain/catalog';
import type {Publication} from '../domain/publication';

export interface Collection extends Genre {description:string;position:number;visible:boolean;titleIds:string[]}
export interface Taxonomy {genres:Genre[];categories:Genre[];collections:Collection[]}
export interface CatalogFilter {category?:string;collection?:string;ageRatings?:string[];hasChapters?:boolean; featured?:boolean; query?: string; kinds?:string[]; genres?:string[]; originalStatuses?: string[]; translationStatuses?: string[]; formats?:string[]; sources?:string[]; countries?:string[]; chaptersMin?:number; chaptersMax?:number; yearMin?:number; yearMax?:number; sort?:string }
export function catalogParams(filter:CatalogFilter):URLSearchParams {
 const params=new URLSearchParams();
 for(const [key,value] of [['category',filter.category],['collection',filter.collection],['q',filter.query]] as const) if(value)params.set(key,value);
 if(filter.sort&&filter.sort!=='updated')params.set('sort',filter.sort);
 if(filter.featured)params.set('featured','true');
 if(filter.hasChapters)params.set('has_chapters','true');
 for(const [key,values] of [['kind',filter.kinds],['age_rating',filter.ageRatings],['genre',filter.genres],['original_status',filter.originalStatuses],['translation_status',filter.translationStatuses],['format',filter.formats],['source',filter.sources],['country',filter.countries]] as const) for(const value of values||[]) if(value)params.append(key,value);
 for(const [key,value] of [['chapters_min',filter.chaptersMin],['chapters_max',filter.chaptersMax],['year_min',filter.yearMin],['year_max',filter.yearMax]] as const) if(value&&value>0)params.set(key,String(value));
 return params;
}
export interface ChapterFilter { query?:string; sort?:'asc'|'desc'; limit?:number }
export interface Genre { id: string; name: string; slug: string; nameRu?:string; categorySlug?:string; parentSlug?:string; position?:number }
export interface Summary {
 ageRating?:AgeRating;
 id: string; slug: string; name: string; kind: Kind; originalStatus: OriginalStatus; translationStatus: TranslationStatus;
 coverURL: string; demo: boolean; chapterCount: number; ratingCount: number; score: number; updatedAt: string;
}
export interface Project { id: string; titleId: string; teamId: string; titleName: string; titleSlug: string; teamName: string; teamSlug: string; status: string }
export interface TitleDetail {bannerURL?:string;categories?:Genre[];collections?:Collection[];ageRating?:AgeRating;listed?:boolean; authors?:string[]; featured?:boolean;
 publication?:Publication; commentCount?:number; chapterCount?:number; readerCount?:number; ratingDistribution?:number[];
 id: string; slug: string; name: string; originalName: string; aliases: string[]; kind: Kind; originalStatus: OriginalStatus; translationStatus: TranslationStatus;
 description: string; direction: 'ltr' | 'rtl'; demo: boolean; coverURL: string; score: number; ratingCount: number; genres: Genre[]; projects: Project[];
}
export interface Release { id: string; projectId: string; teamId: string; teamName: string; teamSlug: string; publishedAt: string }
export interface Chapter { id: string; label: string; volume: string; name: string; order: number; releases: Release[] }
export interface PublicTeam {members?:{name:string;role:string}[]; id: string; name: string; slug: string; description: string;telegramURL?:string; projects: Project[] }
export interface ResultPage<Item> { items: Item[]; nextCursor: string | null }
export interface ReleaseUpdate {releaseId:string;titleId:string;titleName:string;titleSlug:string;chapterLabel:string;teamName:string;teamSlug:string;demo:boolean;publishedAt:string;coverURL?:string}
export interface CatalogGateway {
 rankings(kind?:string):Promise<ResultPage<Summary>>;
 updates(after?:string,team?:string):Promise<ResultPage<ReleaseUpdate>>;
 list(filter: CatalogFilter, cursor?: string): Promise<ResultPage<Summary>>;
 taxonomy?():Promise<Taxonomy>;
 genres(): Promise<Genre[]>;
 detail(slug: string): Promise<TitleDetail | null>;
 chapters(slug: string, cursor?: string, filter?:ChapterFilter): Promise<ResultPage<Chapter>>;
 team(slug: string): Promise<PublicTeam | null>;
}
export async function loadCatalogue(gateway: CatalogGateway, filter: CatalogFilter, cursor?: string) {
 try {
  const [page, taxonomy] = await Promise.all([gateway.list(filter, cursor), gateway.taxonomy?.()]);
  const genres = taxonomy?.genres?.length ? taxonomy.genres : await gateway.genres();
  return { available: true, page, genres, categories:taxonomy?.categories||[],collections:taxonomy?.collections||[], message: '' };
 } catch {
  return { available: false, page: { items: [], nextCursor: null } as ResultPage<Summary>, genres: [] as Genre[],categories:[] as Genre[],collections:[] as Collection[], message: 'Каталогты жүктеу мүмкін болмады. Қайта көріңіз.' };
 }
}
