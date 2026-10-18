import type {ModerationReport} from './community';

export interface ReportQueue {items:ModerationReport[];next:string;hasMore:boolean}

export function receiveReports(queue:ReportQueue,page:ReportQueue,append:boolean):ReportQueue {
 return {...page,items:append?[...queue.items,...page.items]:page.items};
}

export function resolveReport(queue:ReportQueue,id:string):ReportQueue {
 // The cursor describes the fetched page, even when every visible item is resolved.
 return {...queue,items:queue.items.filter(item=>item.id!==id)};
}
