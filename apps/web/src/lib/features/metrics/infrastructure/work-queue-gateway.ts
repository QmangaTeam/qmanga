import type {WorkQueueGateway,QueueKind,WorkPage} from '../application/work-queue';
import type {PublishingGateway} from '$lib/features/publishing/application/publishing';
import type {ProjectRequestsGateway} from '$lib/features/publishing/application/project-requests';
import type {TeamsGateway} from '$lib/features/teams/application/teams';
import type {TitleChangesGateway} from '$lib/features/catalog/application/title-changes';
import type {CommunityGateway} from '$lib/features/community/application/community';
export class ExistingWorkQueueGateway implements WorkQueueGateway {
 constructor(private publishing:PublishingGateway,private changes:TitleChangesGateway,private projects:ProjectRequestsGateway,private teams:TeamsGateway,private community:CommunityGateway){}
 async list(kind:QueueKind,actor:string,cursor=''):Promise<WorkPage>{
  const offset=Math.max(0,Number(cursor)||0);
  if(kind==='publications'){const items=await this.publishing.pending(offset);return {items:items.map(item=>({id:item.id,title:item.titleName,team:item.teamName,description:item.chapterLabel,href:'/admin/moderation?revision='+encodeURIComponent(item.id),pages:item.pages.length||undefined})),next:items.length===100?String(offset+items.length):''};}
  if(kind==='changes'){const page=await this.changes.queue(actor,cursor);return {items:page.items.map(item=>({id:item.id,title:item.title_name,team:item.team_name,description:item.requested_name,href:'/moderation/title-changes/'+encodeURIComponent(item.id)+'?returnTo='+encodeURIComponent('/admin?queue=changes'),createdAt:item.created_at})),next:page.has_more?page.items.at(-1)?.id||'':''};}
  if(kind==='projects'){const page=await this.projects.list(true,cursor);return {items:page.items.filter(item=>item.status==='pending').map(item=>({id:item.id,title:item.titleName||item.name,team:item.teamName,description:item.note,href:'/admin/requests#queue-'+encodeURIComponent(item.id),createdAt:item.createdAt})),next:page.hasMore?page.items.at(-1)?.id||'':''};}
  if(kind==='teams'){const items=await this.teams.pending(offset);return {items:items.map(item=>({id:item.id,title:item.name,team:item.name,description:item.description,href:'/admin/teams#queue-'+encodeURIComponent(item.id)})),next:items.length===100?String(offset+items.length):''};}
  const page=await this.community.reports(actor,cursor);return {items:page.items.filter(item=>item.status==='open').map(item=>({id:item.id,title:item.titleName||item.targetType,team:item.teamName,description:item.reason,href:'/admin/reports#queue-'+encodeURIComponent(item.id),createdAt:item.createdAt})),next:page.hasMore?page.next||page.items.at(-1)?.id||'':''};
 }
}
