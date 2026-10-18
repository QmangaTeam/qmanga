export interface AvatarSource {size:number;type:string}
export interface AvatarGateway {upload(kind:'user'|'team',ownerId:string,source:AvatarSource):Promise<void>}
export function avatarSourceValid(source:AvatarSource):boolean{return source.size>0&&source.size<=5*1024*1024&&['image/jpeg','image/png','image/webp'].includes(source.type);}
