// A local refresh hint after a confirmed upload; the server remains authoritative.
const listeners=new Set<(actorId:string)=>void>();
export const avatarChanges={subscribe(listener:(actorId:string)=>void){listeners.add(listener);return ()=>{listeners.delete(listener);};},publish(actorId:string){for(const listener of listeners)listener(actorId);}};
