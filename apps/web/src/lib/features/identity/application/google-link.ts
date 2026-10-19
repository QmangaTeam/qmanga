export interface GoogleLinkStatus {enabled:boolean;linked:boolean}
export interface GoogleLinkGateway {
 status(accountId:string):Promise<GoogleLinkStatus>;
 start(accountId:string):Promise<string>;
}
