import {validAgeRating, type AgeRating} from '../domain/catalog';
import { ApiFailure, type ApiClient } from '../../../shared/infrastructure/api-client';
import { kinds, originalStatuses, translationStatuses } from '../domain/catalog';
import type { Taxonomy,Collection, ReleaseUpdate, CatalogGateway, CatalogFilter, ChapterFilter, Summary, Genre, Project, Release, Chapter, TitleDetail, ResultPage, PublicTeam } from '../application/catalog';
import type { CatalogManagementGateway } from '../application/management';
import {emptyPublication,validPublication,type Publication} from '../domain/publication';
import type { Metadata } from '../domain/metadata';

function object(value: unknown): Record<string, unknown> {
 if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new Error('Сервер жауабы дұрыс емес');
 return value as Record<string, unknown>;
}
export class HttpCatalogManagementGateway implements CatalogManagementGateway {
 constructor(private api: ApiClient) {}
 async taxonomy(){return taxonomy(await this.api.call('/api/v1/admin/taxonomy'));}
 async saveTaxon(kind:'genres'|'categories',input:Genre){await this.api.call('/api/v1/admin/'+kind+(input.id?'/'+encodeURIComponent(input.id):''),input.id?'PUT':'POST',{name:input.name,slug:input.slug});}
 async deleteTaxon(kind:'genres'|'categories'|'collections',id:string){await this.api.call('/api/v1/admin/'+kind+'/'+encodeURIComponent(id),'DELETE');}
 async saveCollection(input:Collection){await this.api.call('/api/v1/admin/collections'+(input.id?'/'+encodeURIComponent(input.id):''),input.id?'PUT':'POST',{name:input.name,slug:input.slug,description:input.description,position:input.position,visible:input.visible,title_ids:input.titleIds});}
 async setCategories(title:string,ids:string[]){await this.api.call('/api/v1/admin/titles/'+encodeURIComponent(title)+'/categories','PUT',{category_ids:ids});}
 async titles(offset=0) { return page(await this.api.call('/api/v1/admin/titles?offset='+offset), summary).items; }
 async detail(id: string) { return detail(await this.api.call('/api/v1/admin/titles/' + encodeURIComponent(id))); }
 async project(id: string) { return project(await this.api.call('/api/v1/admin/projects/' + encodeURIComponent(id))); }
 async genres() { return page(await this.api.call('/api/v1/genres'), genre).items; }
 async teams(offset=0) { return page(await this.api.call('/api/v1/admin/catalog-teams?offset='+offset), value => { const item = object(value); return { id: text(item.id), name: text(item.name) }; }).items; }
 async save(input: Metadata, id?: string) {
  const result = object(await this.api.call('/api/v1/admin/titles' + (id ? '/' + encodeURIComponent(id) : ''), id ? 'PUT' : 'POST', {...(input.ageRating===undefined?{}:{age_rating:input.ageRating}),publication:input.publication||emptyPublication(),authors:input.authors||[],listed:input.listed===true,featured:input.featured||false,slug: input.slug, name: input.name, original_name: input.originalName, aliases: input.aliases, kind: input.kind, original_status: input.originalStatus, translation_status: input.translationStatus, description: input.description, direction: input.direction, genre_ids: input.genreIds, demo: input.demo }));
  return text(result.id);
 }
 async createGenre(name: string, slug: string) { await this.api.call('/api/v1/admin/genres', 'POST', { name, slug }); }
 async createChapter(titleId: string, label: string, order: number, volume: string, name: string) { await this.api.call('/api/v1/admin/titles/' + encodeURIComponent(titleId) + '/chapters', 'POST', { label, order, volume, name }); }
 async createProject(titleId: string, teamId: string) { await this.api.call('/api/v1/admin/titles/' + encodeURIComponent(titleId) + '/projects', 'POST', { team_id: teamId }); }
}
function text(value: unknown): string { if (typeof value !== 'string') throw new Error('Сервер жауабы дұрыс емес'); return value; }
function number(value: unknown): number { if (typeof value !== 'number' || !Number.isFinite(value)) throw new Error('Сервер жауабы дұрыс емес'); return value; }
function integer(value: unknown): number { const parsed = number(value); if (!Number.isSafeInteger(parsed) || parsed < 0) throw new Error('Сервер жауабы дұрыс емес'); return parsed; }
function boolean(value: unknown): boolean { if (typeof value !== 'boolean') throw new Error('Сервер жауабы дұрыс емес'); return value; }
function choice<Value extends string>(value: unknown, options: readonly Value[]): Value { if (!options.includes(value as Value)) throw new Error('Сервер жауабы дұрыс емес'); return value as Value; }
function array<Item>(value: unknown, map: (entry: unknown) => Item): Item[] { if (!Array.isArray(value)) throw new Error('Сервер жауабы дұрыс емес'); return value.map(map); }
function date(value: unknown): string { const parsed = text(value); if (!Number.isFinite(Date.parse(parsed))) throw new Error('Сервер жауабы дұрыс емес'); return parsed; }
function media(value: unknown): string { const url = text(value); if (url !== '' && !/^\/api\/v1\/media\/[A-Za-z0-9_-]{1,128}$/.test(url)) throw new Error('Сурет сілтемесі жарамсыз'); return url; }
function page<Item>(value: unknown, map: (entry: unknown) => Item): ResultPage<Item> {
 const payload = object(value);
 return { items: array(payload.items, map), nextCursor: payload.next_cursor === null ? null : text(payload.next_cursor) };
}
function genre(value: unknown): Genre { const item = object(value); return { id: text(item.id), name: text(item.name), slug: text(item.slug) }; }
function collection(value:unknown):Collection{const item=object(value);return {...genre(item),description:text(item.description),position:integer(item.position),visible:boolean(item.visible),titleIds:array(item.title_ids,text)};}
function taxonomy(value:unknown):Taxonomy{const item=object(value);return {genres:array(item.genres,genre),categories:array(item.categories,genre),collections:array(item.collections,collection)};}
function project(value: unknown): Project { const item = object(value); return { id: text(item.id), titleId: text(item.title_id), teamId: text(item.team_id), titleName: text(item.title_name), titleSlug: text(item.title_slug), teamName: text(item.team_name), teamSlug: text(item.team_slug), status: text(item.status) }; }
function ageRating(value:unknown):AgeRating { if(value===undefined)return 0; if(!validAgeRating(value))throw new Error('Invalid age rating'); return value; }
function summary(value: unknown): Summary {
 const item = object(value);
 return { ageRating:ageRating(item.age_rating), id: text(item.id), slug: text(item.slug), name: text(item.name), kind: choice(item.kind, kinds), originalStatus: choice(item.original_status, originalStatuses), translationStatus: choice(item.translation_status, translationStatuses), coverURL: media(item.cover_url), demo: boolean(item.demo), chapterCount: integer(item.chapter_count), score: number(item.score), ratingCount: integer(item.rating_count), updatedAt: date(item.updated_at) };
}
function publication(value:unknown):Publication {
 if(value===undefined)return emptyPublication();
 const item=object(value);
 const result:Publication={year:item.year?integer(item.year):0,publishers:array(item.publishers??[],text),source:text(item.source??''),country:text(item.country??''),formats:array(item.formats??[],text),links:array(item.links??[],value=>{const link=object(value);return {label:text(link.label),url:text(link.url)};})};
 if(!validPublication(result))throw new Error('Invalid publication metadata');
 return result;
}
function detail(value: unknown): TitleDetail {
 const item = object(value);
 return {categories:array(item.categories??[],genre),collections:array(item.collections??[],collection), ageRating:ageRating(item.age_rating), id: text(item.id), slug: text(item.slug), name: text(item.name), originalName: text(item.original_name),publication:publication(item.publication),chapterCount:item.chapter_count===undefined?undefined:integer(item.chapter_count),readerCount:item.reader_count===undefined?undefined:integer(item.reader_count),ratingDistribution:item.rating_distribution===undefined?undefined:array(item.rating_distribution,integer), authors:item.authors===undefined?[]:array(item.authors,text),listed:item.listed===true,featured:item.featured===true,aliases: array(item.aliases, text), kind: choice(item.kind, kinds), originalStatus: choice(item.original_status, originalStatuses), translationStatus: choice(item.translation_status, translationStatuses), description: text(item.description), direction: choice(item.direction, ['ltr', 'rtl']), demo: boolean(item.demo), coverURL: media(item.cover_url), score: number(item.score), ratingCount: integer(item.rating_count), genres: array(item.genres, genre), projects: array(item.projects, project) };
}
function release(value: unknown): Release { const item = object(value); return { id: text(item.id), projectId: text(item.project_id), teamId: text(item.team_id), teamName: text(item.team_name), teamSlug: text(item.team_slug), publishedAt: date(item.published_at) }; }
function chapter(value: unknown): Chapter { const item = object(value); return { id: text(item.id), label: text(item.label), volume: text(item.volume), name: text(item.name), order: integer(item.order), releases: array(item.releases, release) }; }
export class HttpCatalogGateway implements CatalogGateway {
 async taxonomy(){return taxonomy(await this.api.call("/api/v1/taxonomy"));}
 constructor(private api: ApiClient) {}
 async rankings(kind='') {
  return page(await this.api.call('/api/v1/rankings?'+new URLSearchParams({kind})),summary);
 }
 async list(filter: CatalogFilter, cursor?: string) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries({ category:filter.category,collection:filter.collection,age_rating:filter.ageRating, q: filter.query, kind: filter.kind, genre: filter.genre, original_status: filter.originalStatus, translation_status: filter.translationStatus, sort: filter.sort, has_chapters:filter.hasChapters?'true':undefined,featured:filter.featured?'true':undefined,cursor })) if (value) query.set(key, value);
  return page(await this.api.call('/api/v1/titles?' + query), summary);
 }
 async genres() { return page(await this.api.call('/api/v1/genres'), genre).items; }
 async detail(slug: string) {
  try { return detail(await this.api.call('/api/v1/titles/' + encodeURIComponent(slug))); }
  catch (failure) { if (failure instanceof ApiFailure && failure.status === 404) return null; throw failure; }
 }
 async updates(after='',team=''):Promise<ResultPage<ReleaseUpdate>> {
  return page(await this.api.call('/api/v1/updates?'+new URLSearchParams({after,team})),value=>{const item=object(value);return {releaseId:text(item.release_id),titleId:text(item.title_id),titleName:text(item.title_name),titleSlug:text(item.title_slug),chapterLabel:text(item.chapter_label),teamName:text(item.team_name),teamSlug:text(item.team_slug),demo:boolean(item.demo),publishedAt:date(item.published_at),coverURL:item.cover_url===undefined?'':media(item.cover_url)};});
 }
 async chapters(slug: string, cursor?: string, filter:ChapterFilter={}) {
  const query=new URLSearchParams();
  if(cursor)query.set('cursor',cursor);
  if(filter.query)query.set('q',filter.query);
  if(filter.sort)query.set('sort',filter.sort);
  if(filter.limit)query.set('limit',String(filter.limit));
  return page(await this.api.call('/api/v1/titles/'+encodeURIComponent(slug)+'/chapters?'+query),chapter);
 }
 async team(slug: string): Promise<PublicTeam | null> {
  try {
   const item = object(await this.api.call('/api/v1/teams/' + encodeURIComponent(slug)));
   return { id: text(item.id), name: text(item.name), slug: text(item.slug), description: text(item.description),telegramURL:typeof item.telegram_url==='string'&&/^https:\/\/t\.me\/[A-Za-z][A-Za-z0-9_]{4,31}$/.test(item.telegram_url)?item.telegram_url:'', members:array(item.members||[],value=>{const member=object(value);return {name:text(member.name),role:text(member.role)};}),projects: array(item.projects, project) };
  } catch (failure) { if (failure instanceof ApiFailure && failure.status === 404) return null; throw failure; }
 }
}
