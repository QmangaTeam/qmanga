export interface RightsInput {titleSlug:string;email:string;reason:string}
export interface RightsGateway {
 create(input:RightsInput):Promise<{id:string;token:string}>;
 status(id:string,token:string):Promise<{status:string;resolution:string}>;
}
export interface PersonalReport {id:string;category:string;reason:string;status:string;resolution:string;createdAt:string}
