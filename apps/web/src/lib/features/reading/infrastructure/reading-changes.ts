import type {ReadingChanges} from '../application/progress';
// A local refresh hint; personal data is always fetched again with account protection.
const listeners=new Set<()=>void>();
export const readingChanges:ReadingChanges={subscribe(listener){listeners.add(listener);return()=>{listeners.delete(listener);};},publish(){for(const listener of listeners)listener();}};
