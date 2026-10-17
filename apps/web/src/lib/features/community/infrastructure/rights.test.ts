import {test,expect} from 'bun:test';
import {HttpRightsGateway} from './http-rights-gateway';
test('rights contact is anonymous and receipt stays in the request body',async()=>{
 const calls:{path:string;init:RequestInit}[]=[];
 const request=(async(path,init)=>{calls.push({path:String(path),init:init!});return Response.json(calls.length===1?{id:'report',token:'secret'}:{id:'report',status:'resolved',resolution:'Checked'});}) as typeof fetch;
 const gateway=new HttpRightsGateway(request);
 expect(await gateway.create({titleSlug:'title',email:'owner@example.test',reason:'Rights request'})).toEqual({id:'report',token:'secret'});
 expect((await gateway.status('report','secret')).resolution).toBe('Checked');
 expect(calls.map(item=>item.path)).toEqual(['/api/v1/rights-reports','/api/v1/rights-reports/status']);
 expect(calls[0].init.credentials).toBe('omit');
 expect(calls[1].init.body).toBe(JSON.stringify({id:'report',token:'secret'}));
});
