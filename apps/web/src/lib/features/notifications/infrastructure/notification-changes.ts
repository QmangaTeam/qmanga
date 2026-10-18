// Local notifications are a refresh hint, never the source of the unread count.
const listeners=new Set<()=>void>();
export const notificationChanges={subscribe(listener:()=>void){listeners.add(listener);return ()=>{listeners.delete(listener);};},publish(){for(const listener of listeners)listener();}};
