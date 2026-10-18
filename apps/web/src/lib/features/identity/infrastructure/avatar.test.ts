import {test,expect} from 'bun:test';
import {HttpAvatarGateway} from './http-avatar-gateway';
test('avatar upload sends rendered account identity even when cookies changed',async()=>{
 const calls:{url:string;init?:RequestInit}[]=[];
 const request=(async(input:RequestInfo|URL,init?:RequestInit)=>{calls.push({url:String(input),init});if(String(input).endsWith('/csrf'))return Response.json({csrf_token:'new-cookie-account-csrf'});return Response.json({url:'/media/avatars/user/other'});}) as typeof fetch;
 const gateway=new HttpAvatarGateway(request);const source=new Blob(['image'],{type:'image/png'});
 await gateway.upload('user','rendered-account',source,'rendered-account');
 expect(calls[1].init?.headers).toEqual({'Content-Type':'image/png','X-CSRF-Token':'new-cookie-account-csrf','X-Qmanga-Account':'rendered-account'});
 await gateway.upload('team','team-id',source,'rendered-team-actor');expect(calls[3].url).toBe('/api/v1/avatars/team/team-id');expect((calls[3].init?.headers as Record<string,string>)['X-Qmanga-Account']).toBe('rendered-team-actor');
});
