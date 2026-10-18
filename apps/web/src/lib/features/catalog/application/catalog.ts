import type { Kind, OriginalStatus, TranslationStatus } from '../domain/catalog';
import type {Publication} from '../domain/publication';

export interface CatalogFilter {hasChapters?:boolean; featured?:boolean; query?: string; kind?: string; genre?: string; originalStatus?: string; translationStatus?: string; sort?: string }
export interface ChapterFilter { query?:string; sort?:'asc'|'desc'; limit?:number }
export interface Genre { id: string; name: string; slug: string }
export interface Summary {
 id: string; slug: string; name: string; kind: Kind; originalStatus: OriginalStatus; translationStatus: TranslationStatus;
 coverURL: string; demo: boolean; chapterCount: number; ratingCount: number; score: number; updatedAt: string;
}
export interface Project { id: string; titleId: string; teamId: string; titleName: string; titleSlug: string; teamName: string; teamSlug: string; status: string }
export interface TitleDetail {listed?:boolean; authors?:string[]; featured?:boolean;
 publication?:Publication; chapterCount?:number; readerCount?:number; ratingDistribution?:number[];
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
 genres(): Promise<Genre[]>;
 detail(slug: string): Promise<TitleDetail | null>;
 chapters(slug: string, cursor?: string, filter?:ChapterFilter): Promise<ResultPage<Chapter>>;
 team(slug: string): Promise<PublicTeam | null>;
}
export async function loadCatalogue(gateway: CatalogGateway, filter: CatalogFilter, cursor?: string) {
 try {
  const [page, genres] = await Promise.all([gateway.list(filter, cursor), gateway.genres()]);
  return { available: true, page, genres, message: '' };
 } catch {
  return { available: false, page: { items: [], nextCursor: null } as ResultPage<Summary>, genres: [] as Genre[], message: 'Каталогты жүктеу мүмкін болмады. Қайта көріңіз.' };
 }
}
