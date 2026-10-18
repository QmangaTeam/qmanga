export type TeamRole = 'member' | 'editor' | 'owner';
export interface Team { id: string; slug: string; name: string; description: string; telegramURL?:string; approved: boolean; reviewStatus?:string; reviewNote?:string; role?: TeamRole }
export interface Member { userId: string; email: string; role: TeamRole }
export interface NewTeam { slug: string; name: string; description: string; telegramURL?:string }
export interface Invitation {id:string;email:string;role:Exclude<TeamRole,'owner'>;status:'pending'|'expired'|'revoked';expiresAt:string}
export function validTelegram(value:string):boolean{return !value||/^https:\/\/t\.me\/[A-Za-z][A-Za-z0-9_]{4,31}\/?$/.test(value);}
export function validTeam(input: NewTeam): boolean {
 return validTelegram(input.telegramURL?.trim()||'') && /^[a-z0-9]+(-[a-z0-9]+)*$/.test(input.slug) && input.slug.length >= 3 && input.slug.length <= 64 && [...input.name.trim()].length >= 2 && [...input.name.trim()].length <= 80 && [...input.description].length <= 2000;
}
export function canManage(team: Team): boolean { return team.role === 'owner'; }
