import {test,expect} from 'bun:test';
import {ApiClient} from '../../../shared/infrastructure/api-client';
import {HttpTeamsGateway} from './http-teams-gateway';
test('team profile writes Telegram and invitations expose only actionable metadata',async()=>{
 const requests:{url:string;method:string;body:any}[]=[];
 const gateway=new HttpTeamsGateway(new ApiClient((async(url,options)=>{
  if(String(url)==='/api/v1/auth/csrf')return Response.json({csrf_token:'test-token'});
  requests.push({url:String(url),method:options?.method||'GET',body:options?.body?JSON.parse(String(options.body)):null});
  if(options?.method==='PUT'||options?.method==='DELETE')return new Response(null,{status:204});
  return Response.json({items:[{id:'invite',email:'member@example.test',role:'editor',status:'revoked',expires_at:'2026-10-20T00:00:00Z',token:'must-not-escape'}]});
 }) as typeof fetch));
 await gateway.update('team',{slug:'dala',name:'Дала',description:'',telegramURL:'https://t.me/dala_team'});
 expect(requests[0].body.telegram_url).toBe('https://t.me/dala_team');expect(requests[0].body.telegramURL).toBeUndefined();
 expect(await gateway.invitations('team',100)).toEqual([{id:'invite',email:'member@example.test',role:'editor',status:'revoked',expiresAt:'2026-10-20T00:00:00Z'}]);
 expect(requests[1].url).toContain('/team/invitations?offset=100');
 await gateway.revokeInvitation('team','invite');expect(requests[2]).toEqual({url:'/api/v1/studio/teams/team/invitations/invite',method:'DELETE',body:null});
});
