export const queueKinds=['publications','changes','projects','teams','reports'] as const;
export type QueueKind=typeof queueKinds[number];
export interface WorkItem {id:string;title:string;team:string;description:string;href:string;createdAt?:string;pages?:number}
export interface WorkPage {items:WorkItem[];next:string}
export interface WorkQueueGateway {list(kind:QueueKind,actor:string,cursor?:string):Promise<WorkPage>}
export function filterWork(items:WorkItem[],query:string,team:string){const term=query.trim().toLocaleLowerCase();return items.filter(item=>(!team||item.team===team)&&(!term||[item.title,item.team,item.description].some(value=>value.toLocaleLowerCase().includes(term))));}
export function loadedCount(page:WorkPage|null){return page?String(page.items.length)+(page.next?'+':''):'—';}

/** A selected row may be beyond the destination's first page. */
export function workPageContext(page:WorkPage,depth:number):WorkPage {return {...page,items:page.items.map(item=>{const url=new URL(item.href,'https://queue.invalid');if(['/admin/requests','/admin/teams','/admin/reports'].includes(url.pathname))url.searchParams.set('loaded',String(depth));return {...item,href:url.pathname+url.search+url.hash};})};}
