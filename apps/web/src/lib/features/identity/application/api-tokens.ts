export interface ApiToken {id:string;name:string;prefix:string;createdAt:string;expiresAt:string;lastUsedAt:string|null;revokedAt:string|null;directPublish:boolean}
export interface ApiTokensGateway {
 list(account:string):Promise<ApiToken[]>;
 create(account:string,input:{name:string;expiresInDays:number;directPublish:boolean}):Promise<{token:string;item:ApiToken}>;
 revoke(account:string,id:string):Promise<void>;
}
export function apiTokenStatus(item:ApiToken,now=Date.now()):'active'|'expired'|'revoked'{return item.revokedAt?'revoked':Date.parse(item.expiresAt)<=now?'expired':'active';}
