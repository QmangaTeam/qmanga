import type { Metadata } from '../domain/metadata';
import type { Genre, Summary, TitleDetail } from './catalog';
export interface CatalogManagementGateway {
 titles(offset?:number): Promise<Summary[]>;
 detail(id: string): Promise<TitleDetail>;
 genres(): Promise<Genre[]>;
 teams(offset?:number): Promise<{ id: string; name: string }[]>;
 save(input: Metadata, id?: string): Promise<string>;
 createGenre(name: string, slug: string): Promise<void>;
 createChapter(titleId: string, label: string, order: number, volume: string, name: string): Promise<void>;
 createProject(titleId: string, teamId: string): Promise<void>;
}
export function editableMetadata(title: TitleDetail): Metadata {
 return { authors:[...(title.authors||[])],featured:title.featured||false,slug: title.slug, name: title.name, originalName: title.originalName, aliases: [...title.aliases], kind: title.kind, originalStatus: title.originalStatus, translationStatus: title.translationStatus, description: title.description, direction: title.direction, genreIds: title.genres.map(genre => genre.id), demo: title.demo };
}
