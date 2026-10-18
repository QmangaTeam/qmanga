export interface ManagedAccount {id:string;name:string;role:string;blocked:boolean;version:number}
export interface AuditEntry {actorName:string;targetName:string;targetPath:string;before:Record<string,string|boolean|number>;after:Record<string,string|boolean|number>;id:string;actor:string;action:string;entityType:string;entityId:string;reason:string;createdAt:string}
export interface ManagementGateway {
 account():Promise<string|null>;
 accounts(account:string,query:string,after?:string):Promise<{items:ManagedAccount[];hasMore:boolean}>;
 change(account:string,target:ManagedAccount,reason:string):Promise<void>;
 audit(account:string,after?:string):Promise<{items:AuditEntry[];hasMore:boolean}>;
}
