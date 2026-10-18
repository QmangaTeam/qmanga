import {test,expect} from 'bun:test';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {HttpCommunityGateway} from './http-community-gateway';

test('moderation response keeps a server cursor and readable target context',async()=>{
 const gateway=new HttpCommunityGateway(new ApiClient((async()=>Response.json({items:[{id:'report',target_type:'comment',target_id:'comment',title_name:'A story',title_slug:'story',team_name:'Team',team_slug:'team',chapter_label:'12',page_position:3,comment_body:'Reported text',comment_author:'Reader',history:[{action:'hide_comment',actor_name:'Moderator',reason:'Spam confirmed',created_at:'2026-10-18'}]}],next:'stable-cursor',has_more:true})) as typeof fetch),{currentId:async()=> 'staff'});
 const page=await gateway.reports('staff');
 expect(page.next).toBe('stable-cursor');
 expect(page.items[0].titleName).toBe('A story');
 expect(page.items[0].commentBody).toBe('Reported text');
 expect(page.items[0].history[0].actorName).toBe('Moderator');
});
