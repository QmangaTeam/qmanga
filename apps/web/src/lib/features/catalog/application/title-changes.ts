import type {Publication} from '../domain/publication';
import type {Kind,OriginalStatus,TranslationStatus,AgeRating} from '../domain/catalog';
import type {Taxonomy} from './catalog';
export interface ProposedTitleData {name:string;original_name:string;aliases:string[];authors:string[];kind:Kind;original_status:OriginalStatus;translation_status:TranslationStatus;description:string;direction:'ltr'|'rtl';age_rating:AgeRating;publication:Publication;genre_ids:string[];category_ids:string[]}
export interface ChangeSnapshot {data:ProposedTitleData;cover_id:string;banner_id:string}
export type ArtworkAction='keep'|'replace'|'remove';
export interface TitleChangeProposal {base_version:string;data:ProposedTitleData;reason:string;links:string[];cover_action:ArtworkAction;banner_action:ArtworkAction;cover_page_id:string}
export interface ChangeAsset {id:string;purpose:'cover'|'banner'|'evidence';url:string;width:number;height:number;position:number}
export interface TitleChange {id:string;project_id:string;title_id:string;title_name:string;team_name:string;requested_name:string;status:'pending'|'approved'|'rejected';reason:string;resolution:string;reviewed_name:string;base_version:string;cover_action:ArtworkAction;banner_action:ArtworkAction;cover_asset_id:string;banner_asset_id:string;links:string[];before:ChangeSnapshot;proposed:ProposedTitleData;assets:ChangeAsset[];created_at:string;reviewed_at:string|null}
export interface TitleChangeContext {project_id:string;title_id:string;title_name:string;team_name:string;slug:string;snapshot:ChangeSnapshot;base_version:string}
export interface TitleChangeDetail {request:TitleChange;current:ChangeSnapshot;current_version:string;can_review:boolean}
export interface ChangeFiles {cover?:File;banner?:File;evidence:File[]}
export interface TitleChangesGateway {
 currentAccount():Promise<string|null>;watchAccount(listener:()=>void):()=>void;
 context(actor:string,project:string):Promise<TitleChangeContext>;
 taxonomy():Promise<Taxonomy>;
 list(actor:string,project:string,after?:string):Promise<{items:TitleChange[];has_more:boolean}>;
 queue(actor:string,after?:string):Promise<{items:TitleChange[];has_more:boolean}>;
 detail(actor:string,id:string):Promise<TitleChangeDetail>;
 submit(actor:string,project:string,proposal:TitleChangeProposal,files:ChangeFiles):Promise<TitleChange>;
 review(actor:string,id:string,decision:'approve'|'reject',reason:string):Promise<void>;
}
