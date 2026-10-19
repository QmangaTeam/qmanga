import {test,expect} from 'bun:test';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {HttpTitleChanges} from './http-title-changes';
const data={name:'Title',original_name:'',aliases:[],authors:[],kind:'manga',original_status:'ongoing',translation_status:'ongoing',description:'',direction:'ltr',age_rating:0,publication:{year:0,publishers:null,source:'',country:'',formats:null,links:null},genre_ids:[],category_ids:[]};
const snapshot={data,cover_id:'',banner_id:''};const version='a'.repeat(64);
const item={id:'request',project_id:'project',title_id:'title',title_name:'Title',team_name:'Team',requested_name:'Editor',status:'pending',reason:'Official source',resolution:'',reviewed_name:'',base_version:version,cover_action:'keep',banner_action:'keep',cover_asset_id:'',banner_asset_id:'',links:['https://example.org'],before:snapshot,proposed:data,assets:[],created_at:'2026-10-19T00:00:00Z',reviewed_at:null};
test('title change gateway binds multipart submission to the form account and normalizes empty publication arrays',async()=>{
 let sent:RequestInit|undefined;
 const api=new ApiClient((async(path:string,init?:RequestInit)=>{if(path.endsWith('/csrf'))return Response.json({csrf_token:'token'});sent=init;return Response.json(item);}) as typeof fetch);
 const gateway=new HttpTitleChanges(api,{currentId:async()=>'different-current-account'});
 const result=await gateway.submit('form-account','project',{base_version:version,data:{...data,publication:{year:0,publishers:[],source:'',country:'',formats:[],links:[]}} as never,reason:'Official source',links:['https://example.org'],cover_action:'keep',banner_action:'keep',cover_page_id:''},{evidence:[new File(['image'],'proof.png',{type:'image/png'})]});
 expect(result.proposed.publication.publishers).toEqual([]);
 expect(new Headers(sent!.headers).get('X-Qmanga-Account')).toBe('form-account');expect(new Headers(sent!.headers).get('Content-Type')).toBeNull();expect(new Headers(sent!.headers).get('X-CSRF-Token')).toBe('token');expect(sent!.body instanceof FormData).toBe(true);expect((sent!.body as FormData).getAll('evidence')).toHaveLength(1);
});
test('title change gateway rejects arbitrary evidence media URLs',async()=>{
 for(const url of ['https://foreign.example/evidence','//foreign.example/image','/api/v1/media/../private']){
 const api=new ApiClient((async()=>Response.json({...item,assets:[{id:'proof',purpose:'evidence',url,width:100,height:100,position:1}]})) as typeof fetch);
 const gateway=new HttpTitleChanges(api,{currentId:async()=>'actor'});
 await expect(gateway.submit('actor','project',{} as never,{evidence:[]})).rejects.toThrow();
 }
});
