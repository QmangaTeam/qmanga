import {test,expect} from 'bun:test';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {HttpManagementGateway} from './http-management-gateway';

test('audit maps readable changes without exposing arbitrary metadata or unsafe destinations',async()=>{
 const gateway=new HttpManagementGateway(new ApiClient((async()=>Response.json({items:[{id:'audit',actor:'staff',actor_name:'Moderator',entity_type:'user',entity_id:'reader',target_name:'Reader',target_path:'/moderation/accounts?q=reader',before:{blocked:false,email:'private@example.test'},after:{blocked:true,token:'secret'}},{id:'unsafe',target_path:'//external.test'}]})) as typeof fetch),{currentId:async()=> 'admin'});
 const result=await gateway.audit('admin');
 expect(result.items[0].actorName).toBe('Moderator');
 expect(result.items[0].targetName).toBe('Reader');
 expect(result.items[0].targetPath).toBe('/moderation/accounts?q=reader');
 expect(result.items[0].before).toEqual({blocked:false});
 expect(result.items[0].after).toEqual({blocked:true});
 expect(result.items[1].targetPath).toBe('');
});

test('cover selection and removal preserve safe asset IDs through audit mapping',async()=>{
 const gateway=new HttpManagementGateway(new ApiClient((async()=>Response.json({items:[{action:'select_cover',before:{cover_asset_id:''},after:{cover_asset_id:'cover-1',object_key:'private'}},{action:'remove_cover',before:{cover_asset_id:'cover-1'},after:{cover_asset_id:''}}]})) as typeof fetch),{currentId:async()=> 'admin'});
 const result=await gateway.audit('admin');
 expect(result.items[0].before).toEqual({cover_asset_id:''});
 expect(result.items[0].after).toEqual({cover_asset_id:'cover-1'});
 expect(result.items[1].before).toEqual({cover_asset_id:'cover-1'});
 expect(result.items[1].after).toEqual({cover_asset_id:''});
});

test('audit forwards all applied filters and cursor to the server',async()=>{let requested='';const gateway=new HttpManagementGateway(new ApiClient((async(input)=>{requested=String(input);return Response.json({items:[]});}) as typeof fetch),{currentId:async()=> 'admin'});await gateway.audit('admin','cursor',{actor:'Ерлан',object:'title',action:'update',from:'2026-10-20T00:00:00Z',to:''});const url=new URL(requested,'https://local.test');expect(url.searchParams.get('actor')).toBe('Ерлан');expect(url.searchParams.get('after')).toBe('cursor');expect(url.searchParams.get('action')).toBe('update');expect(url.searchParams.get('from')).toBe('2026-10-20T00:00:00Z');});
