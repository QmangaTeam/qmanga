import type { Team, Member, NewTeam } from '../domain/team';
import type { TeamsGateway } from '../application/teams';
import type { ApiClient } from '$lib/shared/infrastructure/api-client';
function object(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null; }
function team(value: unknown): Team {
 if (!object(value) || typeof value.id !== 'string' || typeof value.name !== 'string' || typeof value.slug !== 'string' || typeof value.description !== 'string' || typeof value.approved !== 'boolean') throw new Error('Сервер жауабы дұрыс емес');
 const role = value.my_role;
 return { id: value.id, name: value.name, slug: value.slug, description: value.description, approved: value.approved,reviewStatus:typeof value.review_status==='string'?value.review_status:(value.approved?'approved':'pending'),reviewNote:typeof value.review_note==='string'?value.review_note:'', ...(role === 'owner' || role === 'editor' || role === 'member' ? { role } : {}) };
}
function items(value: unknown): unknown[] { if (!object(value) || !Array.isArray(value.items)) throw new Error('Сервер жауабы дұрыс емес'); return value.items; }
export class HttpTeamsGateway implements TeamsGateway {
 constructor(private api: ApiClient) {}
 async reject(id:string,reason:string){await this.api.call('/api/v1/admin/teams/'+encodeURIComponent(id)+'/reject','POST',{reason});}
 async update(id:string,input:NewTeam){await this.api.call('/api/v1/studio/teams/'+encodeURIComponent(id),'PUT',input);}
 async list(offset=0) { return items(await this.api.call('/api/v1/studio/teams?offset='+offset)).map(team); }
 async pending(offset=0) { return items(await this.api.call('/api/v1/admin/teams/pending?offset='+offset)).map(team); }
 async create(input: NewTeam) { return team(await this.api.call('/api/v1/studio/teams', 'POST', input)); }
 async approve(id: string) { await this.api.call('/api/v1/admin/teams/' + encodeURIComponent(id) + '/approve', 'POST', {}); }
 async members(id: string, offset=0): Promise<Member[]> {
  return items(await this.api.call('/api/v1/studio/teams/' + encodeURIComponent(id) + '/members?offset='+offset)).map(value => {
   if (!object(value) || typeof value.user_id !== 'string' || typeof value.email !== 'string' || value.role !== 'member' && value.role !== 'editor' && value.role !== 'owner') throw new Error('Сервер жауабы дұрыс емес');
   return { userId: value.user_id, email: value.email, role: value.role };
  });
 }
 async invite(id: string, email: string, role: 'member' | 'editor') { await this.api.call('/api/v1/studio/teams/' + encodeURIComponent(id) + '/invitations', 'POST', { email, role }); }
 async remove(id: string, userId: string) { await this.api.call('/api/v1/studio/teams/' + encodeURIComponent(id) + '/members/' + encodeURIComponent(userId), 'DELETE'); }
 async accept(token: string) { await this.api.call('/api/v1/teams/invitations/accept', 'POST', { token }); }
}
