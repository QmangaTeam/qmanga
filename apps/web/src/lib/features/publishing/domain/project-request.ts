export type RequestState='pending'|'approved'|'rejected'|'cancelled';
export interface ProjectRequestInput {teamId:string;titleId:string;name:string;originalName:string;slug:string;kind:string;description:string;sourceURL:string;note:string}
export interface ProjectRequest extends ProjectRequestInput {id:string;teamName:string;titleName:string;status:RequestState;resolution:string;projectId:string;createdAt:string}
export function validProjectRequest(input:ProjectRequestInput){return input.teamId!==''&&[...input.note.trim()].length>=8&&[...input.note].length<=2000&&(input.titleId!==''||input.name.trim()!==''&&['manga','manhwa','manhua'].includes(input.kind));}
