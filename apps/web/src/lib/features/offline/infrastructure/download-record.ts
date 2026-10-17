import type {Download} from '../application/downloads';
export function downloadFrom(value:Download):Download {
 return {...value,notice:value.status==='ready'?'none':value.status==='partial'?'download_partial':'download_failed'};
}
