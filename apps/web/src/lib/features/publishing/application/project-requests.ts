import type {ProjectRequest,ProjectRequestInput} from '../domain/project-request';
import {validProjectRequest} from '../domain/project-request';
import {AppFailure} from '$lib/shared/domain/failure';
export interface ProjectRequestsGateway {
 list(admin?:boolean,after?:string,filter?:{team:string;kind:string}):Promise<{items:ProjectRequest[];hasMore:boolean}>;
 create(input:ProjectRequestInput):Promise<void>;
 review(id:string,decision:'approve'|'reject',reason:string,titleId?:string):Promise<void>;
 cancel(id:string):Promise<void>;
 search(query:string):Promise<{id:string;name:string;slug:string}[]>;
}
export function requestTranslation(gateway:ProjectRequestsGateway,input:ProjectRequestInput){if(!validProjectRequest(input))return Promise.reject(new AppFailure('validation'));return gateway.create({...input,note:input.note.trim(),name:input.name.trim()});}
