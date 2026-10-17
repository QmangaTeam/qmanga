export interface Comment {id:string;userId:string;parentId:string;body:string;spoiler:boolean;hidden:boolean;createdAt:string;replies:number}
export interface Discussion {comments:Comment[];next:string;score:number;ratings:number}
export interface Report {titleId:string;releaseId:string;pageId:string;targetType:'title'|'release'|'page'|'comment';targetId:string;category:string;reason:string}
export interface ModerationReport extends Report {id:string;status:string;createdAt:string}
export interface CommunityGateway {
 account():Promise<string|null>;
 discussion(titleId:string,releaseId:string,after?:string):Promise<Discussion>;
 replies(parentId:string,after?:string):Promise<{comments:Comment[];next:string}>;
 comment(accountId:string,titleId:string,releaseId:string,parentId:string,body:string,spoiler:boolean):Promise<void>;
 edit(accountId:string,id:string,body:string,spoiler:boolean):Promise<void>;
 remove(accountId:string,id:string):Promise<void>;
 rating(accountId:string,titleId:string,score:number):Promise<void>;
 myRating(accountId:string,titleId:string):Promise<number>;
 report(accountId:string,input:Report):Promise<void>;
 reports(accountId:string,after?:string):Promise<{items:ModerationReport[];hasMore:boolean}>;
 resolve(accountId:string,id:string,status:string,reason:string):Promise<void>;
 visibility(accountId:string,id:string,hidden:boolean,reason:string):Promise<void>;
}
