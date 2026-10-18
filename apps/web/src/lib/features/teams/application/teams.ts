import {AppFailure} from '../../../shared/domain/failure';
import type { Team, Member, NewTeam, TeamRole,Invitation } from '../domain/team';
import { validTeam } from '../domain/team';
export interface TeamsGateway {
 reject(id:string,reason:string):Promise<void>;
 update(id:string,input:NewTeam):Promise<void>;
 list(offset?:number): Promise<Team[]>;
 create(input: NewTeam): Promise<Team>;
 members(teamId: string, offset?:number): Promise<Member[]>;
 invite(teamId: string, email: string, role: Exclude<TeamRole, 'owner'>): Promise<void>;
 invitations(teamId:string,offset?:number):Promise<Invitation[]>;
 revokeInvitation(teamId:string,id:string):Promise<void>;
 remove(teamId: string, userId: string): Promise<void>;
 accept(token: string): Promise<void>;
 pending(offset?:number): Promise<Team[]>;
 approve(teamId: string): Promise<void>;
}
export function requestTeam(gateway: TeamsGateway, input: NewTeam): Promise<Team> {
 if (!validTeam(input)) return Promise.reject(new AppFailure('team_validation'));
 return gateway.create({ ...input, name: input.name.trim(), description: input.description.trim() });
}
