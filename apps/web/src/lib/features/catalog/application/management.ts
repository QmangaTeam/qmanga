import type { Metadata } from '../domain/metadata';
import type { Genre, Summary, TitleDetail, Project, Taxonomy,Collection } from './catalog';
export interface CatalogManagementGateway {
 taxonomy():Promise<Taxonomy>;
 saveTaxon(kind:'genres'|'categories',input:Genre):Promise<void>;
 deleteTaxon(kind:'genres'|'categories'|'collections',id:string):Promise<void>;
 saveCollection(input:Collection):Promise<void>;
 setCategories(title:string,ids:string[]):Promise<void>;
 titles(offset?:number): Promise<Summary[]>;
 detail(id: string): Promise<TitleDetail>;
 project(id: string): Promise<Project>;
 genres(): Promise<Genre[]>;
 teams(offset?:number): Promise<{ id: string; name: string }[]>;
 save(input: Metadata, id?: string): Promise<string>;
 createGenre(name: string, slug: string): Promise<void>;
 createChapter(titleId: string, label: string, order: number, volume: string, name: string): Promise<void>;
 createProject(titleId: string, teamId: string): Promise<void>;
}
export function editableMetadata(title: TitleDetail): Metadata {
 return {ageRating:title.ageRating??0,publication:title.publication?{...title.publication,publishers:[...title.publication.publishers],formats:[...title.publication.formats],links:title.publication.links.map(link=>({...link}))}:undefined,listed:title.listed===true, authors:[...(title.authors||[])],featured:title.featured||false,slug: title.slug, name: title.name, originalName: title.originalName, aliases: [...title.aliases], kind: title.kind, originalStatus: title.originalStatus, translationStatus: title.translationStatus, description: title.description, direction: title.direction, genreIds: title.genres.map(genre => genre.id), demo: title.demo };
}
