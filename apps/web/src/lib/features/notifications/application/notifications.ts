export interface Notification {id:string;kind:string;title:string;body:string;releaseId:string;titleSlug:string;read:boolean;createdAt:string}
export interface NotificationsGateway{account():Promise<string|null>;list(accountId:string,after?:string):Promise<{items:Notification[];hasMore:boolean}>;read(accountId:string,id:string):Promise<void>}
