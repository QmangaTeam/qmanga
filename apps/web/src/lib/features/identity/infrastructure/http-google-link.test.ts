import {test,expect} from 'bun:test';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {HttpGoogleLinkGateway} from './http-google-link';
test('Google linking status and start remain bound to the rendered account and CSRF',async()=>{
 const requests:{path:string;init?:RequestInit}[]=[];
 const gateway=new HttpGoogleLinkGateway(new ApiClient((async(path:string,init?:RequestInit)=>{requests.push({path,init});if(path.endsWith('/csrf'))return Response.json({csrf_token:'csrf'});if(path.endsWith('/start'))return Response.json({url:'https://accounts.google.com/o/oauth2/v2/auth?state=bound'});return Response.json({enabled:true,linked:false});})as typeof fetch));
 expect(await gateway.status('rendered-account')).toEqual({enabled:true,linked:false});
 expect(await gateway.start('rendered-account')).toBe('https://accounts.google.com/o/oauth2/v2/auth?state=bound');
 const status=requests.find(r=>r.path.endsWith('/link'))!,start=requests.find(r=>r.path.endsWith('/start'))!;
 expect(new Headers(status.init!.headers).get('X-Qmanga-Account')).toBe('rendered-account');
 expect(new Headers(start.init!.headers).get('X-Qmanga-Account')).toBe('rendered-account');expect(new Headers(start.init!.headers).get('X-CSRF-Token')).toBe('csrf');expect(start.init!.method).toBe('POST');
});
test('Google link gateway rejects malformed status and foreign authorization redirects',async()=>{
 for(const value of [{enabled:true,linked:'yes'},{enabled:'yes',linked:false}]){const gateway=new HttpGoogleLinkGateway(new ApiClient((async()=>Response.json(value))as typeof fetch));await expect(gateway.status('actor')).rejects.toThrow();}
 for(const url of ['https://evil.example/oauth','https://accounts.google.com.evil.example/o/oauth2/v2/auth','https://user@accounts.google.com/o/oauth2/v2/auth','http://accounts.google.com/o/oauth2/v2/auth']){const gateway=new HttpGoogleLinkGateway(new ApiClient((async(path:string)=>Response.json(path.endsWith('/csrf')?{csrf_token:'csrf'}:{url}))as typeof fetch));await expect(gateway.start('actor')).rejects.toThrow();}
});
