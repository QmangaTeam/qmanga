import {AppFailure} from '../../../shared/domain/failure';
import type { Team, Member, NewTeam, TeamRole } from '../domain/team';
import { validTeam } from '../domain/team';
export interface TeamsGateway {
 list(): Promise<Team[]>;
 create(input: NewTeam): Promise<Team>;
 members(teamId: string): Promise<Member[]>;
 invite(teamId: string, email: string, role: Exclude<TeamRole, 'owner'>): Promise<void>;
 remove(teamId: string, userId: string): Promise<void>;
 accept(token: string): Promise<void>;
 pending(): Promise<Team[]>;
 approve(teamId: string): Promise<void>;
}
export function requestTeam(gateway: TeamsGateway, input: NewTeam): Promise<Team> {
 if (!validTeam(input)) return Promise.reject(new AppFailure('team_validation'));
 return gateway.create({ ...input, name: input.name.trim(), description: input.description.trim() });
}
